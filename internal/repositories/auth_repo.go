package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type AuthRepository interface {
	CheckPermissionByUid(userUid uint, boardUid uint) bool
	CheckPermissionForAction(userUid uint, action models.UserAction) bool
	CheckRefreshToken(userUid uint, refreshToken string) bool
	ConsumeVerificationCode(verifyUid uint, code string, expectedEmail string) (string, bool)
	ClearRefreshToken(userUid uint)
	FindMyInfoByIDPW(id string, pw string) models.MyInfoResult
	FindMyInfoByUid(userUid uint) models.MyInfoResult
	FindUserInfoByUid(userUid uint) (models.UserInfoResult, error)
	FindUserPasswordByUid(userUid uint) string
	FindUserUidById(id string) uint
	GetAdminUid(boardUid uint) models.BoardAdminUid
	InsertRefreshToken(userUid uint, token string)
	InsertVerificationCode(id string, code string) uint
	SaveRefreshToken(userUid uint, refreshToken string)
	RotateRefreshToken(userUid uint, oldRefreshToken string, newRefreshToken string) bool
	SaveVerificationCode(id string, code string) uint
	DeleteVerificationCode(verifyUid uint)
	UpdateRefreshToken(userUid uint, token string)
	UpdateUserPasswordHash(userUid uint, newBcryptHash string)
	UpdateUserSignin(userUid uint)
	UpdateVerificationCode(id string, code string, uid uint)
}

const verificationCodeLifetime = 10 * time.Minute

type NuboAuthRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboAuthRepository(db *sql.DB) *NuboAuthRepository {
	return &NuboAuthRepository{db: db}
}

// 게시판, 그룹 혹은 최고 관리자인지 확인
func (r *NuboAuthRepository) CheckPermissionByUid(userUid uint, boardUid uint) bool {
	if userUid == 1 {
		return true
	}
	adminUid := r.GetAdminUid(boardUid)
	if userUid == adminUid.Group || userUid == adminUid.Board {
		return true
	}
	return false
}

// 사용자가 지정된 액션에 대한 권한이 있는지 확인
func (r *NuboAuthRepository) CheckPermissionForAction(userUid uint, action models.UserAction) bool {
	query := fmt.Sprintf("SELECT %s AS action FROM %s%s WHERE user_uid = ? LIMIT 1",
		action.String(), configs.Env.Prefix, models.TABLE_USER_PERM)

	var actionValue uint8
	err := r.db.QueryRow(query, userUid).Scan(&actionValue)
	if err == sql.ErrNoRows {
		return true // 별도 기록이 없다면 기본 허용
	}
	return actionValue > 0
}

// 리프레시 토큰이 유효한지 확인
func (r *NuboAuthRepository) CheckRefreshToken(userUid uint, refreshToken string) bool {
	var timestamp int64
	query := fmt.Sprintf("SELECT timestamp FROM %s%s WHERE user_uid = ? AND refresh = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)

	hashed := utils.GetHashedString(refreshToken)
	err := r.db.QueryRow(query, userUid, hashed).Scan(&timestamp)
	if err != nil {
		return false
	}

	_, refreshDays := configs.GetJWTAccessRefresh()
	now := time.Now().UnixMilli()
	validTerm := timestamp + int64(refreshDays*86400000)

	return validTerm > now
}

// 인증 코드를 만료 시간 안에 한 번만 소비한다.
func (r *NuboAuthRepository) ConsumeVerificationCode(verifyUid uint, code string, expectedEmail string) (string, bool) {
	tx, err := r.db.Begin()
	if err != nil {
		return "", false
	}
	defer tx.Rollback()

	var email, storedCode string
	var timestamp int64
	query := fmt.Sprintf("SELECT email, code, timestamp FROM %s%s WHERE uid = ? LIMIT 1 FOR UPDATE",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)
	if err := tx.QueryRow(query, verifyUid).Scan(&email, &storedCode, &timestamp); err != nil {
		return "", false
	}
	if storedCode != code || (expectedEmail != "" && email != expectedEmail) || !verificationCodeValid(timestamp, time.Now()) {
		return "", false
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER_VERIFY)
	result, err := tx.Exec(query, verifyUid)
	if err != nil {
		return "", false
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 || tx.Commit() != nil {
		return "", false
	}
	return email, true
}

func verificationCodeValid(timestamp int64, now time.Time) bool {
	createdAt := time.UnixMilli(timestamp)
	return !createdAt.After(now) && now.Sub(createdAt) <= verificationCodeLifetime
}

// 로그아웃 시 리프레시 토큰 비우기
func (r *NuboAuthRepository) ClearRefreshToken(userUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET refresh = ?, timestamp = ? WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)

	r.db.Exec(query, "", time.Now().UnixMilli(), userUid)
}

// 회원번호에 해당하는 사용자의 공개 정보 반환
func (r *NuboAuthRepository) FindUserInfoByUid(userUid uint) (models.UserInfoResult, error) {
	info := models.UserInfoResult{}
	query := fmt.Sprintf(`SELECT name, profile, level, signature, signup, signin, blocked 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER)

	var blocked uint
	err := r.db.QueryRow(query, userUid).Scan(
		&info.Name, &info.Profile, &info.Level, &info.Signature, &info.Signup, &info.Signin, &blocked)
	if err != nil {
		return info, err
	}

	info.Uid = userUid
	info.Blocked = blocked > 0
	info.Admin = r.CheckPermissionByUid(userUid, models.FAILED)
	return info, nil
}

// 아이디와 (sha256으로 해시된)비밀번호로 내정보 가져오기
func (r *NuboAuthRepository) FindMyInfoByIDPW(id string, pw string) models.MyInfoResult {
	info := models.MyInfoResult{}
	query := fmt.Sprintf(`SELECT uid, name, profile, level, point, signature, signup 
												FROM %s%s WHERE blocked = 0 AND id = ? AND password = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_USER)

	err := r.db.QueryRow(query, id, pw).Scan(&info.Uid, &info.Name, &info.Profile, &info.Level, &info.Point, &info.Signature, &info.Signup)
	if err == sql.ErrNoRows {
		return info
	}

	info.Id = id
	info.Blocked = false
	info.Signin = uint64(time.Now().UnixMilli())
	info.Admin = r.CheckPermissionByUid(info.Uid, models.FAILED)
	return info
}

// 사용자 고유 번호로 내정보 가져오기
func (r *NuboAuthRepository) FindMyInfoByUid(userUid uint) models.MyInfoResult {
	info := models.MyInfoResult{}
	query := fmt.Sprintf(`SELECT uid, id, name, profile, level, point, signature, signup, signin, blocked 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER)

	err := r.db.QueryRow(query, userUid).Scan(&info.Uid, &info.Id, &info.Name, &info.Profile, &info.Level, &info.Point, &info.Signature, &info.Signup, &info.Signin, &info.Blocked)
	if err == sql.ErrNoRows {
		return info
	}
	info.Admin = r.CheckPermissionByUid(info.Uid, models.FAILED)
	return info
}

// 아이디에 해당하는 고유번호 반환
func (r *NuboAuthRepository) FindUserUidById(id string) uint {
	var userUid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)

	err := r.db.QueryRow(query, id).Scan(&userUid)
	if err != nil {
		return models.FAILED
	}
	return userUid
}

// 사용자의 비밀번호(해시값) 반환
func (r *NuboAuthRepository) FindUserPasswordByUid(userUid uint) string {
	var hash string
	query := fmt.Sprintf("SELECT password FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)

	err := r.db.QueryRow(query, userUid).Scan(&hash)
	if err != nil {
		return ""
	}
	return hash
}

// 지정된 게시판 UID로 그룹/게시판 관리자 반환하기
func (r *NuboAuthRepository) GetAdminUid(boardUid uint) models.BoardAdminUid {
	var groupAdminUid, boardAdminUid uint
	query := fmt.Sprintf(`SELECT g.admin_uid, b.admin_uid FROM %s%s AS g JOIN %s%s AS b 
												ON g.uid = b.group_uid WHERE b.uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_GROUP, configs.Env.Prefix, models.TABLE_BOARD)

	r.db.QueryRow(query, boardUid).Scan(&groupAdminUid, &boardAdminUid)
	return models.BoardAdminUid{Group: groupAdminUid, Board: boardAdminUid}
}

// 사용자의 리프레시 토큰 추가하기
func (r *NuboAuthRepository) InsertRefreshToken(userUid uint, token string) {
	query := fmt.Sprintf("INSERT INTO %s%s (user_uid, refresh, timestamp) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)

	r.db.Exec(query, userUid, token, time.Now().UnixMilli())
}

// 인증코드 추가하기
func (r *NuboAuthRepository) InsertVerificationCode(id string, code string) uint {
	query := fmt.Sprintf("INSERT INTO %s%s (email, code, timestamp) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)

	result, err := r.db.Exec(query, id, code, time.Now().UnixMilli())
	if err != nil {
		return models.FAILED
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 로그인 시 리프레시 토큰 저장하기
func (r *NuboAuthRepository) SaveRefreshToken(userUid uint, refreshToken string) {
	now := time.Now().UnixMilli()
	hashed := utils.GetHashedString(refreshToken)
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)

	var uid uint
	err := r.db.QueryRow(query, userUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %s%s (user_uid, refresh, timestamp) VALUES (?, ?, ?)",
			configs.Env.Prefix, models.TABLE_USER_TOKEN)
		r.db.Exec(query, userUid, hashed, now)
	} else {
		r.UpdateRefreshToken(userUid, hashed)
	}
}

// 저장된 기존 토큰이 아직 유효할 때만 새 토큰으로 원자적으로 교체한다.
func (r *NuboAuthRepository) RotateRefreshToken(userUid uint, oldRefreshToken string, newRefreshToken string) bool {
	_, refreshDays := configs.GetJWTAccessRefresh()
	oldHash := utils.GetHashedString(oldRefreshToken)
	newHash := utils.GetHashedString(newRefreshToken)
	now := time.Now().UnixMilli()
	validSince := now - int64(refreshDays)*24*60*60*1000
	query := fmt.Sprintf(`UPDATE %s%s SET refresh = ?, timestamp = ?
		WHERE user_uid = ? AND refresh = ? AND timestamp > ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER_TOKEN)
	result, err := r.db.Exec(query, newHash, now, userUid, oldHash, validSince)
	if err != nil {
		return false
	}
	rows, err := result.RowsAffected()
	return err == nil && rows == 1
}

// (회원가입 시) 인증 코드 보관해놓기
func (r *NuboAuthRepository) SaveVerificationCode(id string, code string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE email = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)

	err := r.db.QueryRow(query, id).Scan(&uid)
	if err == sql.ErrNoRows {
		return r.InsertVerificationCode(id, code)
	}
	if err != nil {
		return models.FAILED
	}
	r.UpdateVerificationCode(id, code, uid)
	return uid
}

func (r *NuboAuthRepository) DeleteVerificationCode(verifyUid uint) {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER_VERIFY)
	r.db.Exec(query, verifyUid)
}

// 사용자의 리프레시 토큰 업데이트하기
func (r *NuboAuthRepository) UpdateRefreshToken(userUid uint, token string) {
	query := fmt.Sprintf("UPDATE %s%s SET refresh = ?, timestamp = ? WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)

	r.db.Exec(query, token, time.Now().UnixMilli(), userUid)
}

// 인증코드 업데이트하기
func (r *NuboAuthRepository) UpdateVerificationCode(id string, code string, uid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET code = ?, timestamp = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)

	r.db.Exec(query, code, time.Now().UnixMilli(), uid)
}

// 로그인 시간 업데이트
func (r *NuboAuthRepository) UpdateUserSignin(userUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET signin = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)

	r.db.Exec(query, time.Now().UnixMilli(), userUid)
}

// 사용자의 비밀번호를 SHA256 해시값에서 Bcrypt 해시값으로 업데이트
func (r *NuboAuthRepository) UpdateUserPasswordHash(userUid uint, newBcryptHash string) {
	query := fmt.Sprintf("UPDATE %s%s SET password = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)
	r.db.Exec(query, newBcryptHash, userUid)
}

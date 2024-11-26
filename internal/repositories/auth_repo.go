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
	CheckAdminPermission(table models.Table, userUid uint) bool
	CheckPermissionByUid(userUid uint, boardUid uint) bool
	CheckPermissionForAction(userUid uint, action models.UserAction) bool
	CheckVerificationCode(param models.VerifyParameter) bool
	ClearRefreshToken(userUid uint)
	FindUserInfoByUid(userUid uint) (models.UserInfoResult, error)
	FindMyInfoByIDPW(id string, pw string) models.MyInfoResult
	FindMyInfoByUid(userUid uint) models.MyInfoResult
	FindIDCodeByVerifyUid(verifyUid uint) (string, string)
	FindUserUidById(id string) uint
	InsertRefreshToken(userUid uint, token string)
	InsertVerificationCode(id string, code string) uint
	SaveVerificationCode(id string, code string) uint
	SaveRefreshToken(userUid uint, refreshToken string)
	UpdateRefreshToken(userUid uint, token string)
	UpdateVerificationCode(id string, code string, uid uint)
	UpdateUserSignin(userUid uint)
}

type TsboardAuthRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardAuthRepository(db *sql.DB) *TsboardAuthRepository {
	return &TsboardAuthRepository{db: db}
}

// 관리자 권한 확인하기
func (r *TsboardAuthRepository) CheckAdminPermission(table models.Table, userUid uint) bool {
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE admin_uid = ? LIMIT 1", configs.Env.Prefix, table)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	var uid uint
	err = stmt.QueryRow(userUid).Scan(&uid)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		return false
	}
	return true
}

// 게시판, 그룹 혹은 최고 관리자인지 확인
func (r *TsboardAuthRepository) CheckPermissionByUid(userUid uint, boardUid uint) bool {
	if userUid == 1 {
		return true
	}
	if isGroupAdmin := r.CheckAdminPermission(models.TABLE_GROUP, userUid); isGroupAdmin {
		return true
	}
	if boardUid > 0 {
		if isBoardAdmin := r.CheckAdminPermission(models.TABLE_BOARD, userUid); isBoardAdmin {
			return true
		}
	}
	return false
}

// 사용자가 지정된 액션에 대한 권한이 있는지 확인
func (r *TsboardAuthRepository) CheckPermissionForAction(userUid uint, action models.UserAction) bool {
	query := fmt.Sprintf("SELECT %s AS action FROM %s%s WHERE user_uid = ? LIMIT 1",
		action.String(), configs.Env.Prefix, models.TABLE_USER_PERM)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	var actionValue uint8
	err = stmt.QueryRow(userUid).Scan(&actionValue)
	if err == sql.ErrNoRows {
		return true // 별도 기록이 없다면 기본 허용
	}
	return actionValue > 0
}

// 인증 코드가 유효한지 확인
func (r *TsboardAuthRepository) CheckVerificationCode(param models.VerifyParameter) bool {
	var code string
	var timestamp uint64

	query := fmt.Sprintf("SELECT code, timestamp FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false
	}
	defer stmt.Close()

	err = stmt.QueryRow(param.Target).Scan(&code, &timestamp)
	if err == sql.ErrNoRows {
		return false
	}

	now := uint64(time.Now().UnixMilli())
	gap := uint64(1000 * 60 * 10)
	if now > timestamp+gap {
		return false
	}

	if code == param.Code {
		return true
	}
	return false
}

// 로그아웃 시 리프레시 토큰 비우기
func (r *TsboardAuthRepository) ClearRefreshToken(userUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET refresh = ?, timestamp = ? WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	stmt.Exec("", time.Now().UnixMilli(), userUid)
}

// 회원번호에 해당하는 사용자의 공개 정보 반환
func (r *TsboardAuthRepository) FindUserInfoByUid(userUid uint) (models.UserInfoResult, error) {
	info := models.UserInfoResult{}
	query := fmt.Sprintf(`SELECT name, profile, level, signature, signup, signin, blocked 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return info, err
	}
	defer stmt.Close()

	var blocked uint
	err = stmt.QueryRow(userUid).Scan(
		&info.Name, &info.Profile, &info.Level, &info.Signature, &info.Signup, &info.Signin, &blocked)
	if err != nil {
		return info, err
	}

	info.Blocked = blocked > 0
	info.Admin = r.CheckPermissionByUid(userUid, models.FAILED)
	return info, nil
}

// 아이디와 (sha256으로 해시된)비밀번호로 내정보 가져오기
func (r *TsboardAuthRepository) FindMyInfoByIDPW(id string, pw string) models.MyInfoResult {
	info := models.MyInfoResult{}
	query := fmt.Sprintf(`SELECT uid, name, profile, level, point, signature, signup 
												FROM %s%s WHERE blocked = 0 AND id = ? AND password = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return info
	}
	defer stmt.Close()

	err = stmt.QueryRow(id, pw).Scan(&info.Uid, &info.Name, &info.Profile, &info.Level, &info.Point, &info.Signature, &info.Signup)
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
func (r *TsboardAuthRepository) FindMyInfoByUid(userUid uint) models.MyInfoResult {
	info := models.MyInfoResult{}
	query := fmt.Sprintf(`SELECT uid, id, name, profile, level, point, signature, signup, signin, blocked 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return info
	}
	defer stmt.Close()

	err = stmt.QueryRow(userUid).Scan(&info.Uid, &info.Id, &info.Name, &info.Profile, &info.Level, &info.Point, &info.Signature, &info.Signup, &info.Signin, &info.Blocked)
	if err == sql.ErrNoRows {
		return info
	}
	info.Admin = r.CheckPermissionByUid(info.Uid, models.FAILED)
	return info
}

// 인증용 고유번호로 아이디와 코드 가져오기
func (r *TsboardAuthRepository) FindIDCodeByVerifyUid(verifyUid uint) (string, string) {
	var id, code string
	query := fmt.Sprintf("SELECT email, code FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return id, code
	}
	defer stmt.Close()

	stmt.QueryRow(verifyUid).Scan(&id, &code)
	return id, code
}

// 아이디에 해당하는 고유번호 반환
func (r *TsboardAuthRepository) FindUserUidById(id string) uint {
	var userUid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return models.FAILED
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&userUid)
	if err != nil {
		return models.FAILED
	}
	return userUid
}

// 로그인 시 리프레시 토큰 저장하기
func (r *TsboardAuthRepository) SaveRefreshToken(userUid uint, refreshToken string) {
	now := time.Now().UnixMilli()
	hashed := utils.GetHashedString(refreshToken)
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	var uid uint
	err = stmt.QueryRow(userUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %s%s (user_uid, refresh, timestamp) VALUES (?, ?, ?)",
			configs.Env.Prefix, models.TABLE_USER_TOKEN)
		r.db.Exec(query, userUid, hashed, now)
	} else {
		r.UpdateRefreshToken(userUid, hashed)
	}
}

// 사용자의 리프레시 토큰 추가하기
func (r *TsboardAuthRepository) InsertRefreshToken(userUid uint, token string) {
	query := fmt.Sprintf("INSERT INTO %s%s (user_uid, refresh, timestamp) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	stmt.Exec(userUid, token, time.Now().UnixMilli())
}

// 인증코드 추가하기
func (r *TsboardAuthRepository) InsertVerificationCode(id string, code string) uint {
	query := fmt.Sprintf("INSERT INTO %s%s (email, code, timestamp) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return models.FAILED
	}
	defer stmt.Close()

	result, _ := stmt.Exec(id, code, time.Now().UnixMilli())
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// (회원가입 시) 인증 코드 보관해놓기
func (r *TsboardAuthRepository) SaveVerificationCode(id string, code string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE email = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return models.FAILED
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&uid)
	if err == sql.ErrNoRows {
		return r.InsertVerificationCode(id, code)
	}
	r.UpdateVerificationCode(id, code, uid)
	return uid
}

// 사용자의 리프레시 토큰 업데이트하기
func (r *TsboardAuthRepository) UpdateRefreshToken(userUid uint, token string) {
	query := fmt.Sprintf("UPDATE %s%s SET refresh = ?, timestamp = ? WHERE user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_TOKEN)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	stmt.Exec(token, time.Now().UnixMilli(), userUid)
}

// 인증코드 업데이트하기
func (r *TsboardAuthRepository) UpdateVerificationCode(id string, code string, uid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET code = ?, timestamp = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER_VERIFY)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	stmt.Exec(code, time.Now().UnixMilli(), uid)
}

// 로그인 시간 업데이트
func (r *TsboardAuthRepository) UpdateUserSignin(userUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET signin = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	stmt.Exec(time.Now().UnixMilli(), userUid)
}

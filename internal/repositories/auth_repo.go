package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type AuthRepository interface {
	CheckVerificationCode(param *models.VerifyParameter) bool
	CheckPermissionByUid(userUid uint, boardUid uint) bool
	CheckPermissionForAction(userUid uint, action models.Action) bool
	ClearRefreshToken(userUid uint)
	FindUserInfoByUid(userUid uint) (*models.UserInfoResult, error)
	FindMyInfoByIDPW(id string, pw string) *models.MyInfoResult
	FindMyInfoByUid(userUid uint) *models.MyInfoResult
	FindIDCodeByVerifyUid(verifyUid uint) (string, string)
	FindUserUidById(id string) uint
	SaveVerificationCode(id string, code string) uint
	SaveRefreshToken(userUid uint, refreshToken string)
	UpdateUserSignin(userUid uint)
}

type MySQLAuthRepository struct {
	db *sql.DB
}

// *sql.DB 저장
func NewMySQLAuthRepository(db *sql.DB) *MySQLAuthRepository {
	return &MySQLAuthRepository{db: db}
}

// 인증 코드가 유효한지 확인
func (r *MySQLAuthRepository) CheckVerificationCode(param *models.VerifyParameter) bool {
	var code string
	var timestamp uint64

	query := fmt.Sprintf("SELECT code, timestamp FROM %suser_verification WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	err := r.db.QueryRow(query, param.Target).Scan(&code, &timestamp)
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

// 게시판, 그룹 혹은 최고 관리자인지 확인 (boardUid = 0 일 때는 게시판 관리자인지 검사 안함)
func (r *MySQLAuthRepository) CheckPermissionByUid(userUid uint, boardUid uint) bool {
	if userUid == 1 {
		return true
	}

	query := fmt.Sprintf("SELECT uid FROM %sgroup WHERE admin_uid = ? LIMIT 1", configs.Env.Prefix)
	var uid uint
	err := r.db.QueryRow(query, userUid).Scan(&uid)
	if err == sql.ErrNoRows {
		return false
	}

	if boardUid > 0 {
		query = fmt.Sprintf("SELECT uid FROM %sboard WHERE admin_uid = ? LIMIT 1", configs.Env.Prefix)
		err = r.db.QueryRow(query, userUid).Scan(&uid)
		if err == sql.ErrNoRows {
			return false
		}
	}

	return true
}

// 사용자가 지정된 액션에 대한 권한이 있는지 확인
func (r *MySQLAuthRepository) CheckPermissionForAction(userUid uint, action models.Action) bool {
	query := fmt.Sprintf("SELECT %s AS action FROM %suser_permission WHERE user_uid = ? LIMIT 1",
		action.String(), configs.Env.Prefix)

	var actionValue uint8
	err := r.db.QueryRow(query, userUid).Scan(&actionValue)
	if err == sql.ErrNoRows {
		return true // 별도 기록이 없다면 기본 허용
	}
	if actionValue > 0 {
		return true
	}
	return false
}

// 로그아웃 시 리프레시 토큰 비우기
func (r *MySQLAuthRepository) ClearRefreshToken(userUid uint) {
	query := fmt.Sprintf("UPDATE %suser_token SET refresh = ?, timestamp = ? WHERE user_uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.Exec(query, "", time.Now().UnixMilli(), userUid)
}

// 회원번호에 해당하는 사용자의 공개 정보 반환
func (r *MySQLAuthRepository) FindUserInfoByUid(userUid uint) (*models.UserInfoResult, error) {
	query := fmt.Sprintf("SELECT name, profile, level, signature, signup, signin, blocked FROM %suser WHERE uid = ? LIMIT 1", configs.Env.Prefix)

	var blocked uint
	var info models.UserInfoResult

	err := r.db.QueryRow(query, userUid).Scan(
		&info.Name, &info.Profile, &info.Level, &info.Signature, &info.Signup, &info.Signin, &blocked)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal("Failed to execute query: ", err)
		}
		return &info, err
	}

	info.Blocked = blocked > 0
	info.Admin = r.CheckPermissionByUid(userUid, NO_BOARD_UID)

	return &info, nil
}

// 아이디와 (sha256으로 해시된)비밀번호로 내정보 가져오기
func (r *MySQLAuthRepository) FindMyInfoByIDPW(id string, pw string) *models.MyInfoResult {
	query := fmt.Sprintf(`SELECT uid, name, profile, level, point, signature, signup FROM %suser
	 WHERE blocked = 0 AND id = ? AND password = ? LIMIT 1`, configs.Env.Prefix)

	var info models.MyInfoResult
	err := r.db.QueryRow(query, id, pw).Scan(&info.Uid, &info.Name, &info.Profile, &info.Level, &info.Point, &info.Signature, &info.Signup)

	if err == sql.ErrNoRows {
		return &models.MyInfoResult{}
	}

	info.Id = id
	info.Blocked = false
	info.Signin = uint64(time.Now().UnixMilli())
	info.Admin = r.CheckPermissionByUid(info.Uid, NO_BOARD_UID)
	return &info
}

// 사용자 고유 번호로 내정보 가져오기
func (r *MySQLAuthRepository) FindMyInfoByUid(userUid uint) *models.MyInfoResult {
	query := fmt.Sprintf(`SELECT uid, id, name, profile, level, point, signature, signup, signin, blocked FROM %suser
	 WHERE uid = ? LIMIT 1`, configs.Env.Prefix)

	var info models.MyInfoResult
	err := r.db.QueryRow(query, userUid).Scan(&info.Uid, &info.Id, &info.Name, &info.Profile, &info.Level, &info.Point, &info.Signature, &info.Signup, &info.Signin, &info.Blocked)

	if err == sql.ErrNoRows {
		return &models.MyInfoResult{}
	}

	info.Admin = r.CheckPermissionByUid(info.Uid, NO_BOARD_UID)
	return &info
}

// 인증용 고유번호로 아이디와 코드 가져오기
func (r *MySQLAuthRepository) FindIDCodeByVerifyUid(verifyUid uint) (string, string) {
	var id, code string
	query := fmt.Sprintf("SELECT email, code FROM %suser_verification WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.QueryRow(query, verifyUid).Scan(&id, &code)
	return id, code
}

// 아이디에 해당하는 고유번호 반환
func (r *MySQLAuthRepository) FindUserUidById(id string) uint {
	var userUid uint
	query := fmt.Sprintf("SELECT uid FROM %suser WHERE id = ? LIMIT 1", configs.Env.Prefix)
	err := r.db.QueryRow(query, id).Scan(&userUid)
	if err != nil {
		return NOT_FOUND
	}
	return userUid
}

// 로그인 시 리프레시 토큰 저장하기
func (r *MySQLAuthRepository) SaveRefreshToken(userUid uint, refreshToken string) {
	now := time.Now().UnixMilli()
	hashed := utils.GetHashedString(refreshToken)
	query := fmt.Sprintf("SELECT user_uid FROM %suser_token WHERE user_uid = ? LIMIT 1", configs.Env.Prefix)

	var uid uint
	err := r.db.QueryRow(query, userUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %suser_token (user_uid, refresh, timestamp) VALUES (?, ?, ?)", configs.Env.Prefix)
		r.db.Exec(query, userUid, hashed, now)
	} else {
		query = fmt.Sprintf("UPDATE %suser_token SET refresh = ?, timestamp = ? WHERE user_uid = ? LIMIT 1", configs.Env.Prefix)
		r.db.Exec(query, hashed, now, userUid)
	}
}

// (회원가입 시) 인증 코드 보관해놓기
func (r *MySQLAuthRepository) SaveVerificationCode(id string, code string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %suser_verification WHERE email = ? LIMIT 1", configs.Env.Prefix)
	err := r.db.QueryRow(query, id).Scan(&uid)
	now := time.Now().UnixMilli()

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %suser_verification (email, code, timestamp) VALUES (?, ?, ?)", configs.Env.Prefix)
		result, _ := r.db.Exec(query, id, code, now)
		insertId, err := result.LastInsertId()
		if err != nil {
			uid = NOT_FOUND
		}
		uid = uint(insertId)
	} else {
		query = fmt.Sprintf("UPDATE %suser_verification SET code = ?, timestamp = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
		r.db.Exec(query, code, now, uid)
	}
	return uid
}

// 로그인 시간 업데이트
func (r *MySQLAuthRepository) UpdateUserSignin(userUid uint) {
	query := fmt.Sprintf("UPDATE %suser SET signin = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.Exec(query, time.Now().UnixMilli(), userUid)
}

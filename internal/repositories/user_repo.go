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

type UserRepository interface {
	FindUserInfoByUid(userUid uint) (*models.UserInfoResult, error)
	CheckPermissionByUid(userUid uint, boardUid uint) bool
	CheckPermissionForAction(userUid uint, action models.Action) bool
	InsertBlackList(actorUid uint, targetUid uint)
	InsertReportUser(actorUid uint, targetUid uint, report string)
	FindMyInfoByIDPW(id string, pw string) *models.MyInfoResult
	UpdateUserSignin(userUid uint)
	SaveRefreshToken(userUid uint, refreshToken string)
	IsEmailDuplicated(id string) bool
	IsNameDuplicated(name string) bool
	InsertNewUser(id string, pw string, name string) uint
	SaveVerificationCode(id string, code string) uint
	CheckVerificationCode(param *models.VerifyParameter) bool
}

type MySQLUserRepository struct {
	db *sql.DB
}

// *sql.DB 저장
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

const NO_BOARD_UID = 0

// 회원번호에 해당하는 사용자의 공개 정보 반환
func (r *MySQLUserRepository) FindUserInfoByUid(userUid uint) (*models.UserInfoResult, error) {
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

// 게시판, 그룹 혹은 최고 관리자인지 확인 (boardUid = 0 일 때는 게시판 관리자인지 검사 안함)
func (r *MySQLUserRepository) CheckPermissionByUid(userUid uint, boardUid uint) bool {
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
func (r *MySQLUserRepository) CheckPermissionForAction(userUid uint, action models.Action) bool {
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

// 다른 사용자를 내 블랙리스트에 등록하기
func (r *MySQLUserRepository) InsertBlackList(actorUid uint, targetUid uint) {
	query := fmt.Sprintf("SELECT user_uid FROM %suser_black_list WHERE user_uid = ? AND black_uid = ? LIMIT 1",
		configs.Env.Prefix)

	var uid uint
	err := r.db.QueryRow(query, actorUid, targetUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %suser_black_list (user_uid, black_uid) VALUES (?, ?)", configs.Env.Prefix)
		r.db.Exec(query, actorUid, targetUid)
	}
}

// 다른 사용자를 신고하기
func (r *MySQLUserRepository) InsertReportUser(actorUid uint, targetUid uint, report string) {
	query := fmt.Sprintf("SELECT uid FROM %sreport WHERE to_uid = ? AND from_uid = ? LIMIT 1", configs.Env.Prefix)

	var uid uint
	err := r.db.QueryRow(query, targetUid, actorUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %sreport (to_uid, from_uid, request, response, timestamp, solved) VALUES (?, ?, ?, ? ,? ,?)", configs.Env.Prefix)
		r.db.Exec(query, targetUid, actorUid, report, "", time.Now().UnixMilli(), 0)
	}
}

// 아이디와 (sha256으로 해시된)비밀번호로 사용자 고유 번호 반환
func (r *MySQLUserRepository) FindMyInfoByIDPW(id string, pw string) *models.MyInfoResult {
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

// 로그인 시간 업데이트
func (r *MySQLUserRepository) UpdateUserSignin(userUid uint) {
	query := fmt.Sprintf("UPDATE %suser SET signin = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.Exec(query, time.Now().UnixMilli(), userUid)
}

// 로그인 시 리프레시 토큰 저장하기
func (r *MySQLUserRepository) SaveRefreshToken(userUid uint, refreshToken string) {
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

// (회원가입 시) 이메일 주소가 중복되는지 확인
func (r *MySQLUserRepository) IsEmailDuplicated(id string) bool {
	query := fmt.Sprintf("SELECT uid FROM %suser WHERE id = ? LIMIT 1", configs.Env.Prefix)
	var uid uint
	r.db.QueryRow(query, id).Scan(&uid)
	return uid > 0
}

// (회원가입 시) 이름이 중복되는지 확인
func (r *MySQLUserRepository) IsNameDuplicated(name string) bool {
	query := fmt.Sprintf("SELECT uid FROM %suser WHERE name = ? LIMIT 1", configs.Env.Prefix)
	var uid uint
	r.db.QueryRow(query, name).Scan(&uid)
	return uid > 0
}

// 신규 회원 등록
func (r *MySQLUserRepository) InsertNewUser(id string, pw string, name string) uint {
	isDupId := r.IsEmailDuplicated(id)
	isDupName := r.IsNameDuplicated(name)
	if isDupId || isDupName {
		return 0
	}

	query := fmt.Sprintf(`INSERT INTO %suser 
	(id, name, password, profile, level, point, signature, signup, signin, blocked)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix)
	result, err := r.db.Exec(query, id, name, pw, "", 1, 100, "", time.Now().UnixMilli(), 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return 0
	}
	return uint(insertId)
}

// (회원가입 시) 인증 코드 보관해놓기
func (r *MySQLUserRepository) SaveVerificationCode(id string, code string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %suser_verification WHERE email = ? LIMIT 1", configs.Env.Prefix)
	err := r.db.QueryRow(query, id).Scan(&uid)
	now := time.Now().UnixMilli()

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %suser_verification (email, code, timestamp) VALUES (?, ?, ?)", configs.Env.Prefix)
		result, _ := r.db.Exec(query, id, code, now)
		insertId, err := result.LastInsertId()
		if err != nil {
			uid = 0
		}
		uid = uint(insertId)
	} else {
		query = fmt.Sprintf("UPDATE %suser_verification SET code = ?, timestamp = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
		r.db.Exec(query, code, now, uid)
	}
	return uid
}

// 인증 코드가 유효한지 확인
func (r *MySQLUserRepository) CheckVerificationCode(param *models.VerifyParameter) bool {
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

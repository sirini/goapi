package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type UserRepository interface {
	GetReportResponse(userUid uint) string
	InsertBlackList(actorUid uint, targetUid uint)
	InsertReportUser(actorUid uint, targetUid uint, report string)
	InsertNewUser(id string, pw string, name string) uint
	InsertNewChat(fromUserUid uint, toUserUid uint, message string) uint
	IsEmailDuplicated(id string) bool
	IsNameDuplicated(name string) bool
	IsBlocked(userUid uint) bool
	LoadUserPermission(userUid uint) *models.UserPermissionResult
	UpdatePassword(userUid uint, password string)
	UpdateUserInfoString(userUid uint, name string, signature string)
	UpdateUserProfile(userUid uint, imagePath string)
}

type MySQLUserRepository struct {
	db *sql.DB
}

// *sql.DB 저장
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

const NO_BOARD_UID = 0
const NOT_FOUND = 0

// 사용자 신고 내용에 대한 응답 가져오기
func (r *MySQLUserRepository) GetReportResponse(userUid uint) string {
	var response string
	query := fmt.Sprintf("SELECT response FROM %sreport WHERE to_uid = ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix)
	r.db.QueryRow(query, userUid).Scan(&response)
	return response
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

// 신규 회원 등록
func (r *MySQLUserRepository) InsertNewUser(id string, pw string, name string) uint {
	isDupId := r.IsEmailDuplicated(id)
	isDupName := r.IsNameDuplicated(name)
	if isDupId || isDupName {
		return NOT_FOUND
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
		return NOT_FOUND
	}
	return uint(insertId)
}

// 쪽지 보내기
func (r *MySQLUserRepository) InsertNewChat(fromUserUid uint, toUserUid uint, message string) uint {
	query := fmt.Sprintf("INSERT INTO %schat (to_uid, from_uid, message, timestamp) VALUES (?, ?, ?, ?)", configs.Env.Prefix)
	result, _ := r.db.Exec(query, toUserUid, fromUserUid, message, time.Now().UnixMilli())
	insertId, err := result.LastInsertId()
	if err != nil {
		return NOT_FOUND
	}
	return uint(insertId)
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

// 로그인이 차단되었는지 확인
func (r *MySQLUserRepository) IsBlocked(userUid uint) bool {
	var blocked uint8
	query := fmt.Sprintf("SELECT blocked FROM %suser WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.QueryRow(query, userUid).Scan(&blocked)
	return blocked > 0
}

// 사용자 권한 및 신고 받은 후 조치사항 조회
func (r *MySQLUserRepository) LoadUserPermission(userUid uint) *models.UserPermissionResult {
	var result models.UserPermissionResult = models.UserPermissionResult{
		WritePost:       true,
		WriteComment:    true,
		SendChatMessage: true,
		SendReport:      true,
	}

	var writePost, writeComment, sendChat, sendReport uint8
	query := fmt.Sprintf("SELECT write_post, write_comment, send_chat, send_report FROM %suser_permission WHERE user_uid = ? LIMIT 1", configs.Env.Prefix)
	err := r.db.QueryRow(query, userUid).Scan(&writePost, &writeComment, &sendChat, &sendReport)
	if err == sql.ErrNoRows {
		return &result
	}

	result.WritePost = writePost > 0
	result.WriteComment = writeComment > 0
	result.SendChatMessage = sendChat > 0
	result.SendReport = sendReport > 0
	return &result
}

// 사용자 비밀번호 변경하기
func (r *MySQLUserRepository) UpdatePassword(userUid uint, pw string) {
	query := fmt.Sprintf("UPDATE %suser SET password = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.Exec(query, pw, userUid)
}

// 사용자 이름, 서명 변경하기
func (r *MySQLUserRepository) UpdateUserInfoString(userUid uint, name string, signature string) {
	query := fmt.Sprintf("UPDATE %suser SET name = ?, signature = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.Exec(query, name, signature, userUid)
}

// 사용자 프로필 이미지 변경하기
func (r *MySQLUserRepository) UpdateUserProfile(userUid uint, imagePath string) {
	query := fmt.Sprintf("UPDATE %suser SET profile = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix)
	r.db.Exec(query, imagePath, userUid)
}

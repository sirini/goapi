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
	FindUserInfoByUid(userUid uint) (*models.UserInfo, error)
	CheckPermissionByUid(userUid uint, boardUid uint) bool
	CheckPermissionForAction(userUid uint, action models.Action) bool
	InsertBlackList(actorUid uint, targetUid uint)
	InsertReportUser(actorUid uint, targetUid uint, report string)
}

type MySQLUserRepository struct {
	db *sql.DB
}

// *sql.DB 저장
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

// 회원번호에 해당하는 사용자의 공개 정보 반환
func (r *MySQLUserRepository) FindUserInfoByUid(userUid uint) (*models.UserInfo, error) {
	query := fmt.Sprintf("SELECT name, profile, level, signature, signup, signin, blocked FROM %suser WHERE uid = ? LIMIT 1", configs.Env.DBTablePrefix)

	var blocked uint
	var info models.UserInfo

	err := r.db.QueryRow(query, userUid).Scan(
		&info.Name, &info.Profile, &info.Level, &info.Signature, &info.Signup, &info.Signin, &blocked)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal("Failed to execute query: ", err)
		}
		return &info, err
	}

	info.Blocked = blocked > 0
	info.Admin = r.CheckPermissionByUid(userUid, 0)

	return &info, nil
}

// 게시판, 그룹 혹은 최고 관리자인지 확인 (boardUid = 0 일 때는 게시판 관리자인지 검사 안함)
func (r *MySQLUserRepository) CheckPermissionByUid(userUid uint, boardUid uint) bool {
	if userUid == 1 {
		return true
	}

	query := fmt.Sprintf("SELECT uid FROM %sgroup WHERE admin_uid = ? LIMIT 1", configs.Env.DBTablePrefix)

	var uid uint
	err := r.db.QueryRow(query, userUid).Scan(&uid)
	if err == sql.ErrNoRows {
		return false
	}

	if boardUid > 0 {
		query = fmt.Sprintf("SELECT uid FROM %sboard WHERE admin_uid = ? LIMIT 1", configs.Env.DBTablePrefix)
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
		action.String(), configs.Env.DBTablePrefix)

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
		configs.Env.DBTablePrefix)

	var uid uint
	err := r.db.QueryRow(query, actorUid, targetUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %suser_black_list (user_uid, black_uid) VALUES (?, ?)", configs.Env.DBTablePrefix)
		r.db.Exec(query, actorUid, targetUid)
	}
}

// 다른 사용자를 신고하기
func (r *MySQLUserRepository) InsertReportUser(actorUid uint, targetUid uint, report string) {
	query := fmt.Sprintf("SELECT uid FROM %sreport WHERE to_uid = ? AND from_uid = ? LIMIT 1", configs.Env.DBTablePrefix)

	var uid uint
	err := r.db.QueryRow(query, targetUid, actorUid).Scan(&uid)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %sreport (to_uid, from_uid, request, response, timestamp, solved) VALUES (?, ?, ?, ? ,? ,?)", configs.Env.DBTablePrefix)
		r.db.Exec(query, targetUid, actorUid, report, "", time.Now().UnixMilli(), 0)
	}
}

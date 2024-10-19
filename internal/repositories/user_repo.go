package repositories

import (
	"database/sql"
	"log"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type UserRepository interface {
	FindUserInfoByUid(userUid uint) (*models.UserInfo, error)
	CheckPermissionByUid(userUid uint, boardUid uint) bool
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

	query := strings.Join([]string{
		"SELECT name, profile, level, signature, signup, signin, blocked FROM ",
		configs.Env.DBTablePrefix,
		"user WHERE uid = ? LIMIT 1"}, "")

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

	query := strings.Join([]string{
		"SELECT uid FROM ", configs.Env.DBTablePrefix, "group WHERE admin_uid = ? LIMIT 1"}, "")
	err := r.db.QueryRow(query, userUid)
	if err != nil {
		return false
	}

	if boardUid > 0 {
		query = strings.Join([]string{
			"SELECT uid FROM ", configs.Env.DBTablePrefix, "board WHERE admin_uid = ? LIMIT 1"}, "")
		err = r.db.QueryRow(query, userUid)
		if err != nil {
			return false
		}
	}

	return true
}

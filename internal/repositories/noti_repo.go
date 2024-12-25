package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type NotiRepository interface {
	FindBoardUidByPostUidForLoop(stmt *sql.Stmt, postUid uint) uint
	GetBoardIdTypeForLoop(stmt *sql.Stmt, boardUint uint) (string, models.Board)
	GetUserNameProfileForLoop(stmt *sql.Stmt, userUid uint) (string, string)
	InsertNotification(param models.InsertNotificationParameter)
	IsNotiAdded(param models.InsertNotificationParameter) bool
	LoadNotification(userUid uint, limit uint) ([]models.NotificationItem, error)
	UpdateAllChecked(userUid uint, limit uint)
}

type TsboardNotiRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardNotiRepository(db *sql.DB) *TsboardNotiRepository {
	return &TsboardNotiRepository{db: db}
}

// 게시판 고유 번호 가져오기
func (r *TsboardNotiRepository) FindBoardUidByPostUidForLoop(stmt *sql.Stmt, postUid uint) uint {
	var boardUid uint
	stmt.QueryRow(postUid).Scan(&boardUid)
	return boardUid
}

// 게시판 아이디와 타입 가져오기
func (r *TsboardNotiRepository) GetBoardIdTypeForLoop(stmt *sql.Stmt, boardUid uint) (string, models.Board) {
	var id string
	var boardType models.Board
	stmt.QueryRow(boardUid).Scan(&id, &boardType)
	return id, boardType
}

// 사용자의 이름과 프로필 이미지 가져오기
func (r *TsboardNotiRepository) GetUserNameProfileForLoop(stmt *sql.Stmt, userUid uint) (string, string) {
	var name, profile string
	stmt.QueryRow(userUid).Scan(&name, &profile)
	return name, profile
}

// 새 알림 추가하기
func (r *TsboardNotiRepository) InsertNotification(param models.InsertNotificationParameter) {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(to_uid, from_uid, type, post_uid, comment_uid, checked, timestamp)
												VALUES (?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_NOTI)

	r.db.Exec(query, param.TargetUserUid, param.ActionUserUid, param.NotiType, param.PostUid, param.CommentUid, 0, time.Now().UnixMilli())
}

// 중복 알림인지 확인
func (r *TsboardNotiRepository) IsNotiAdded(param models.InsertNotificationParameter) bool {
	query := fmt.Sprintf(`SELECT uid FROM %s%s WHERE to_uid = ? AND from_uid = ?
												AND type = ? AND post_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_NOTI)

	var uid uint
	r.db.QueryRow(query, param.TargetUserUid, param.ActionUserUid, param.NotiType, param.PostUid).Scan(&uid)
	return uid > 0
}

// 나에게 온 알림들 가져오기
func (r *TsboardNotiRepository) LoadNotification(userUid uint, limit uint) ([]models.NotificationItem, error) {
	query := fmt.Sprintf(`SELECT uid, from_uid, type, post_uid, checked, timestamp 
												FROM %s%s WHERE to_uid = ? ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_NOTI)

	rows, err := r.db.Query(query, userUid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 게시판 고유 번호 가져오는 쿼리 준비
	query = fmt.Sprintf("SELECT board_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)
	stmtPost, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtPost.Close()

	// 게시판 아이디, 타입 가져오는 쿼리 준비
	query = fmt.Sprintf("SELECT id, type FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD)
	stmtBoard, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtBoard.Close()

	// 사용자 이름과 프로필 가져오는 쿼리 준비
	query = fmt.Sprintf("SELECT name, profile FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmtUser, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtUser.Close()

	items := make([]models.NotificationItem, 0)
	for rows.Next() {
		item := models.NotificationItem{}
		var checked uint8
		err = rows.Scan(&item.Uid, &item.FromUser.UserUid, &item.Type, &item.PostUid, &checked, &item.Timestamp)
		if err != nil {
			return nil, err
		}
		item.Checked = checked > 0

		boardUid := r.FindBoardUidByPostUidForLoop(stmtPost, item.PostUid)
		if boardUid > 0 {
			item.Id, item.BoardType = r.GetBoardIdTypeForLoop(stmtBoard, boardUid)
		}
		item.FromUser.Name, item.FromUser.Profile = r.GetUserNameProfileForLoop(stmtUser, userUid)
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return items, nil
}

// 모든 알람 확인하기
func (r *TsboardNotiRepository) UpdateAllChecked(userUid uint, limit uint) {
	query := fmt.Sprintf("UPDATE %s%s SET checked = ? WHERE to_uid = ? LIMIT ?",
		configs.Env.Prefix, models.TABLE_NOTI)

	r.db.Exec(query, 1, userUid, limit)
}

package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type NotiRepository interface {
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
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var items []models.NotificationItem
	for rows.Next() {
		item := models.NotificationItem{}
		var checked uint8
		err = rows.Scan(&item.Uid, &item.FromUser.UserUid, &item.Type, &item.PostUid, &checked, &item.Timestamp)
		if err != nil {
			return nil, err
		}
		item.Checked = checked > 0

		var boardUid uint
		query = fmt.Sprintf("SELECT board_uid FROM %s%s WHERE uid = ? LIMIT 1",
			configs.Env.Prefix, models.TABLE_POST)
		r.db.QueryRow(query, item.PostUid).Scan(&boardUid)
		if boardUid > 0 {
			query = fmt.Sprintf("SELECT id, type FROM %s%s WHERE uid = ? LIMIT 1",
				configs.Env.Prefix, models.TABLE_BOARD)
			r.db.QueryRow(query, boardUid).Scan(&item.Id, &item.BoardType)
		}

		query = fmt.Sprintf("SELECT name, profile FROM %s%s WHERE uid = ? LIMIT 1",
			configs.Env.Prefix, models.TABLE_USER)
		r.db.QueryRow(query, item.FromUser.UserUid).Scan(&item.FromUser.Name, &item.FromUser.Profile)
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

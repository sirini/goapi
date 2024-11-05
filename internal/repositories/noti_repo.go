package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type NotiRepository interface {
	InsertNewNotification(param *models.NewNotiParameter)
	IsNotiAdded(param *models.NewNotiParameter) bool
	LoadNotification(userUid uint, limit uint) ([]*models.NotificationItem, error)
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
func (r *TsboardNotiRepository) InsertNewNotification(param *models.NewNotiParameter) {
	query := fmt.Sprintf(`INSERT INTO %snotification 
												(to_uid, from_uid, type, post_uid, comment_uid, checked, timestamp)
												VALUES (?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix)
	r.db.Exec(query, param.TargetUserUid, param.ActionUserUid, param.NotiType, param.PostUid, param.CommentUid, 0, time.Now().UnixMilli())
}

// 중복 알림인지 확인
func (r *TsboardNotiRepository) IsNotiAdded(param *models.NewNotiParameter) bool {
	query := fmt.Sprintf(`SELECT uid FROM %snotification WHERE to_uid = ? AND from_uid = ?
												AND type = ? AND post_uid = ? LIMIT 1`, configs.Env.Prefix)
	var uid uint
	r.db.QueryRow(query, param.TargetUserUid, param.ActionUserUid, param.NotiType, param.PostUid).Scan(&uid)
	return uid > 0
}

// 나에게 온 알림들 가져오기
func (r *TsboardNotiRepository) LoadNotification(userUid uint, limit uint) ([]*models.NotificationItem, error) {
	query := fmt.Sprintf(`SELECT uid, from_uid, type, post_uid, checked, timestamp 
												FROM %snotification WHERE to_uid = ? ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix)
	rows, err := r.db.Query(query, userUid, limit)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var NotiItems []*models.NotificationItem
	for rows.Next() {
		item := &models.NotificationItem{}
		var checked uint8
		err = rows.Scan(&item.Uid, &item.FromUser.UserUid, &item.Type, &item.PostUid, &checked, &item.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		item.Checked = checked > 0

		var boardUid uint
		query = fmt.Sprintf("SELECT board_uid FROM %spost WHERE uid = ? LIMIT 1", configs.Env.Prefix)
		r.db.QueryRow(query, item.PostUid).Scan(&boardUid)
		if boardUid > 0 {
			query = fmt.Sprintf("SELECT id, type FROM %sboard WHERE uid = ? LIMIT 1", configs.Env.Prefix)
			r.db.QueryRow(query, boardUid).Scan(&item.Id, &item.BoardType)
		}

		query = fmt.Sprintf("SELECT name, profile FROM %suser WHERE uid = ? LIMIT 1", configs.Env.Prefix)
		r.db.QueryRow(query, item.FromUser.UserUid).Scan(&item.FromUser.Name, &item.FromUser.Profile)

		NotiItems = append(NotiItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return NotiItems, nil
}

// 모든 알람 확인하기
func (r *TsboardNotiRepository) UpdateAllChecked(userUid uint, limit uint) {
	query := fmt.Sprintf("UPDATE %snotification SET checked = ? WHERE to_uid = ? LIMIT ?", configs.Env.Prefix)
	r.db.Exec(query, 1, userUid, limit)
}

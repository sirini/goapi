package repositories

import (
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type ChatRepository interface {
	InsertNewChat(actionUserUid uint, targetUserUid uint, message string) uint
	LoadChatList(userUid uint, limit uint) ([]models.ChatItem, error)
	LoadChatHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]models.ChatHistory, error)
	MarkChatRead(actionUserUid uint, targetUserUid uint, throughUid uint, readAt uint64) (int64, error)
}

type NuboChatRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboChatRepository(db *sql.DB) *NuboChatRepository {
	return &NuboChatRepository{db: db}
}

// 쪽지 보내기
func (r *NuboChatRepository) InsertNewChat(actionUserUid uint, targetUserUid uint, message string) uint {
	query := fmt.Sprintf("INSERT INTO %s%s (to_uid, from_uid, message, timestamp) VALUES (?, ?, ?, ?)",
		configs.Env.Prefix, models.TABLE_CHAT)

	result, err := r.db.Exec(query, targetUserUid, actionUserUid, message, time.Now().UnixMilli())
	if err != nil {
		return models.FAILED
	}

	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 쪽지 목록들 반환
func (r *NuboChatRepository) LoadChatList(userUid uint, limit uint) ([]models.ChatItem, error) {
	query := chatListQuery(configs.Env.Prefix)

	rows, err := r.db.Query(query, userUid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.ChatItem, 0)
	for rows.Next() {
		item := models.ChatItem{}
		err = rows.Scan(&item.Uid, &item.Sender.UserUid, &item.Message, &item.Timestamp, &item.Sender.Name, &item.Sender.Profile)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.Reverse(items)
	return items, nil
}

func chatListQuery(prefix string) string {
	return fmt.Sprintf(`SELECT c.uid, c.from_uid, c.message, c.timestamp, u.name, u.profile
		FROM %s%s AS c
		JOIN (
			SELECT from_uid, MAX(uid) AS latest_uid
			FROM %s%s WHERE to_uid = ? GROUP BY from_uid
		) AS latest ON latest.latest_uid = c.uid
		JOIN %suser AS u ON c.from_uid = u.uid
		ORDER BY c.uid DESC LIMIT ?`, prefix, models.TABLE_CHAT, prefix, models.TABLE_CHAT, prefix)
}

// 상대방과의 대화 내용 가져오기
func (r *NuboChatRepository) LoadChatHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]models.ChatHistory, error) {
	query := chatHistoryQuery(configs.Env.Prefix)

	rows, err := r.db.Query(query, targetUserUid, actionUserUid, actionUserUid, targetUserUid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.ChatHistory, 0)
	for rows.Next() {
		history := models.ChatHistory{}
		if err := rows.Scan(&history.Uid, &history.UserUid, &history.Message, &history.Timestamp, &history.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, history)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.Reverse(items)
	return items, nil
}

func chatHistoryQuery(prefix string) string {
	return fmt.Sprintf(`SELECT uid, from_uid, message, timestamp, read_at FROM %s%s
		WHERE (to_uid = ? AND from_uid = ?) OR (to_uid = ? AND from_uid = ?)
		ORDER BY uid DESC LIMIT ?`, prefix, models.TABLE_CHAT)
}

// 현재 사용자가 상대방에게서 받은 쪽지만 마지막으로 표시한 지점까지 읽음 처리한다.
func (r *NuboChatRepository) MarkChatRead(actionUserUid uint, targetUserUid uint, throughUid uint, readAt uint64) (int64, error) {
	query := chatReadQuery(configs.Env.Prefix)
	result, err := r.db.Exec(query, readAt, actionUserUid, targetUserUid, throughUid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func chatReadQuery(prefix string) string {
	return fmt.Sprintf(`UPDATE %s%s SET read_at = ?
		WHERE to_uid = ? AND from_uid = ? AND uid <= ? AND read_at = 0`, prefix, models.TABLE_CHAT)
}

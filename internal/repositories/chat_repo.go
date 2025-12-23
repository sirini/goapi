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
	query := fmt.Sprintf(`SELECT MAX(c.uid) AS latest_uid, c.from_uid, MAX(c.message) AS latest_message, 
												MAX(c.timestamp) AS latest_timestamp, u.name, u.profile 
												FROM %s%s AS c JOIN %suser AS u ON c.from_uid = u.uid WHERE c.to_uid = ? 
												GROUP BY c.from_uid ORDER BY latest_uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_CHAT, configs.Env.Prefix)

	rows, err := r.db.Query(query, userUid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chatItems := make([]models.ChatItem, 0)
	for rows.Next() {
		item := models.ChatItem{}
		err = rows.Scan(&item.Uid, &item.Sender.UserUid, &item.Message, &item.Timestamp, &item.Sender.Name, &item.Sender.Profile)
		if err != nil {
			return nil, err
		}
		chatItems = append(chatItems, item)
	}
	slices.Reverse(chatItems)
	return chatItems, nil
}

// 상대방과의 대화 내용 가져오기
func (r *NuboChatRepository) LoadChatHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]models.ChatHistory, error) {
	query := fmt.Sprintf(`SELECT uid, from_uid, message, timestamp FROM %s%s 
												WHERE (to_uid = ? AND from_uid = ?) OR (to_uid = ? AND from_uid = ?)
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_CHAT)

	rows, err := r.db.Query(query, targetUserUid, actionUserUid, actionUserUid, targetUserUid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chatHistories := make([]models.ChatHistory, 0)
	for rows.Next() {
		history := models.ChatHistory{}
		if err := rows.Scan(&history.Uid, &history.UserUid, &history.Message, &history.Timestamp); err != nil {
			return nil, err
		}
		chatHistories = append(chatHistories, history)
	}
	slices.Reverse(chatHistories)
	return chatHistories, nil
}

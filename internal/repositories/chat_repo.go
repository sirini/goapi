package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type ChatRepository interface {
	InsertNewChat(actionUserUid uint, targetUserUid uint, message string) uint
	LoadChatList(userUid uint, limit uint) ([]*models.ChatItem, error)
	LoadChatHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]*models.ChatHistory, error)
}

type TsboardChatRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardChatRepository(db *sql.DB) *TsboardChatRepository {
	return &TsboardChatRepository{db: db}
}

// 쪽지 보내기
func (r *TsboardChatRepository) InsertNewChat(actionUserUid uint, targetUserUid uint, message string) uint {
	query := fmt.Sprintf("INSERT INTO %schat (to_uid, from_uid, message, timestamp) VALUES (?, ?, ?, ?)", configs.Env.Prefix)
	result, _ := r.db.Exec(query, targetUserUid, actionUserUid, message, time.Now().UnixMilli())
	insertId, err := result.LastInsertId()
	if err != nil {
		return NOT_FOUND
	}
	return uint(insertId)
}

// 쪽지 목록들 반환
func (r *TsboardChatRepository) LoadChatList(userUid uint, limit uint) ([]*models.ChatItem, error) {
	query := fmt.Sprintf(`SELECT MAX(c.uid) AS latest_uid, c.from_uid, MAX(c.message) AS latest_message, 
												MAX(c.timestamp) AS latest_timestamp, u.name, u.profile 
		FROM %schat AS c JOIN %suser AS u ON c.from_uid = u.uid WHERE c.to_uid = ? 
		GROUP BY c.from_uid ORDER BY latest_uid DESC LIMIT ?`, configs.Env.Prefix, configs.Env.Prefix)

	rows, err := r.db.Query(query, userUid, limit)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var chatItems []*models.ChatItem
	for rows.Next() {
		item := &models.ChatItem{}
		err = rows.Scan(&item.Uid, &item.Sender.UserUid, &item.Message, &item.Timestamp, &item.Sender.Name, &item.Sender.Profile)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		chatItems = append(chatItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return chatItems, nil
}

// 상대방과의 대화 내용 가져오기
func (r *TsboardChatRepository) LoadChatHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]*models.ChatHistory, error) {
	query := fmt.Sprintf(`SELECT uid, from_uid, message, timestamp FROM %schat 
												WHERE to_uid IN (?, ?) AND from_uid IN (?, ?) 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix)
	rows, err := r.db.Query(query, actionUserUid, targetUserUid, actionUserUid, targetUserUid, limit)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var chatHistories []*models.ChatHistory
	for rows.Next() {
		history := &models.ChatHistory{}
		if err := rows.Scan(&history.Uid, &history.UserUid, &history.Message, &history.Timestamp); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		chatHistories = append(chatHistories, history)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return chatHistories, nil
}

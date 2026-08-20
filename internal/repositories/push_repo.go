package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirini/goapi/internal/configs"
)

type PushRepository interface {
	SaveDevice(userUid uint, token string, platform string) error
	RemoveDevice(userUid uint, token string) error
	FindTokens(userUid uint) ([]string, error)
	RemoveDevices(tokens []string) error
}

func (r *NuboPushRepository) RemoveDevices(tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(tokens)), ",")
	query := fmt.Sprintf("DELETE FROM %spush_device WHERE token IN (%s)", configs.Env.Prefix, placeholders)
	args := make([]any, len(tokens))
	for index, token := range tokens {
		args[index] = token
	}
	_, err := r.db.Exec(query, args...)
	return err
}

type NuboPushRepository struct {
	db *sql.DB
}

func NewNuboPushRepository(db *sql.DB) *NuboPushRepository {
	return &NuboPushRepository{db: db}
}

// 같은 FCM 토큰이 다른 계정으로 로그인되면 현재 계정 소유로 안전하게 이동한다.
func (r *NuboPushRepository) SaveDevice(userUid uint, token string, platform string) error {
	query := fmt.Sprintf(`INSERT INTO %spush_device (user_uid, token, platform, updated)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE user_uid = VALUES(user_uid), platform = VALUES(platform), updated = VALUES(updated)`,
		configs.Env.Prefix)
	_, err := r.db.Exec(query, userUid, token, platform, time.Now().UnixMilli())
	return err
}

func (r *NuboPushRepository) RemoveDevice(userUid uint, token string) error {
	query := fmt.Sprintf("DELETE FROM %spush_device WHERE user_uid = ? AND token = ?", configs.Env.Prefix)
	_, err := r.db.Exec(query, userUid, token)
	return err
}

func (r *NuboPushRepository) FindTokens(userUid uint) ([]string, error) {
	query := fmt.Sprintf("SELECT token FROM %spush_device WHERE user_uid = ? ORDER BY updated DESC", configs.Env.Prefix)
	rows, err := r.db.Query(query, userUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

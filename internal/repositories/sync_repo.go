package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type SyncRepository interface {
	GetFileName(fileUid uint) string
}

type TsboardSyncRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardSyncRepository(db *sql.DB) *TsboardSyncRepository {
	return &TsboardSyncRepository{db: db}
}

// 첨부 파일의 원래 파일명 가져오기
func (r *TsboardSyncRepository) GetFileName(fileUid uint) string {
	var name string
	query := fmt.Sprintf("SELECT name FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE)
	r.db.QueryRow(query, fileUid).Scan(&name)
	return name
}

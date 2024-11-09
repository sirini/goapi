package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type FileRepository interface {
	GetCoverImage(postUid uint) string
}

type TsboardFileRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardFileRepository(db *sql.DB) *TsboardFileRepository {
	return &TsboardFileRepository{db: db}
}

// 게시글 대표 커버 썸네일 이미지 가져오기
func (r *TsboardFileRepository) GetCoverImage(postUid uint) string {
	var path string
	query := fmt.Sprintf("SELECT path FROM %s%s WHERE post_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)
	r.db.QueryRow(query, postUid).Scan(&path)
	return path
}

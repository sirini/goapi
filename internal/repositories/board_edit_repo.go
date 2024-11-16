package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardEditRepository interface {
	InsertImagePath(boardUid uint, userUid uint, paths []string)
}

type TsboardBoardEditRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardBoardEditRepository(db *sql.DB, board BoardRepository) *TsboardBoardEditRepository {
	return &TsboardBoardEditRepository{db: db, board: board}
}

// 게시글에 삽입한 이미지 정보들을 한 번에 저장하기
func (r *TsboardBoardEditRepository) InsertImagePath(boardUid uint, userUid uint, paths []string) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, user_uid, path, timestamp) VALUES `,
		configs.Env.Prefix, models.TABLE_IMAGE)
	values := make([]interface{}, 0)
	now := time.Now().UnixMilli()

	for _, path := range paths {
		query += "(?, ?, ?, ?),"
		values = append(values, boardUid, userUid, path, now)
	}

	query = query[:len(query)-1]
	r.db.Exec(query, values...)
}

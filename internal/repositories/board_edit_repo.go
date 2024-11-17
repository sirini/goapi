package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardEditRepository interface {
	GetInsertedImages(param models.EditorInsertImageParameter) ([]models.Pair, error)
	GetMaxImageUid(boardUid uint, actionUserUid uint) uint
	GetTotalImageCount(boardUid uint, actionUserUid uint) uint
	InsertImagePath(boardUid uint, userUid uint, paths []string)
	RemoveInsertedImage(imageUid uint, actionUserUid uint) string
}

type TsboardBoardEditRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardBoardEditRepository(db *sql.DB, board BoardRepository) *TsboardBoardEditRepository {
	return &TsboardBoardEditRepository{db: db, board: board}
}

// 게시글에 삽입했던 이미지들 가져오기
func (r *TsboardBoardEditRepository) GetInsertedImages(param models.EditorInsertImageParameter) ([]models.Pair, error) {
	var images []models.Pair
	if param.LastUid < 1 {
		param.LastUid = r.GetMaxImageUid(param.BoardUid, param.UserUid) + 1
	}
	query := fmt.Sprintf(`SELECT uid, path FROM %s%s WHERE uid < ? AND board_uid = ? AND user_uid = ? 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_IMAGE)
	rows, err := r.db.Query(query, param.LastUid, param.BoardUid, param.UserUid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		image := models.Pair{}
		rows.Scan(&image.Uid, &image.Name)
		images = append(images, image)
	}
	return images, nil
}

// 내가 올린 이미지들의 가장 최근 고유 번호 반환
func (r *TsboardBoardEditRepository) GetMaxImageUid(boardUid uint, actionUserUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT MAX(uid) FROM %s%s WHERE board_uid = ? AND user_uid = ?",
		configs.Env.Prefix, models.TABLE_IMAGE)
	r.db.QueryRow(query, boardUid, actionUserUid).Scan(&uid)
	return uid
}

// 내가 올린 이미지 총 갯수 반환
func (r *TsboardBoardEditRepository) GetTotalImageCount(boardUid uint, actionUserUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE board_uid = ? AND user_uid = ?",
		configs.Env.Prefix, models.TABLE_IMAGE)
	r.db.QueryRow(query, boardUid, actionUserUid).Scan(&count)
	return count
}

// 게시글에 삽입한 이미지 정보들을 한 번에 저장하기
func (r *TsboardBoardEditRepository) InsertImagePath(boardUid uint, userUid uint, paths []string) {
	query := fmt.Sprintf("INSERT INTO %s%s (board_uid, user_uid, path, timestamp) VALUES ",
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

// 게시글에 삽입한 이미지 삭제하기
func (r *TsboardBoardEditRepository) RemoveInsertedImage(imageUid uint, actionUserUid uint) string {
	query := fmt.Sprintf("SELECT user_uid, path FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_IMAGE)
	var userUid uint
	var path string
	r.db.QueryRow(query, imageUid).Scan(&userUid, &path)

	if actionUserUid != userUid {
		return ""
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE)
	r.db.Exec(query, imageUid)
	return path
}

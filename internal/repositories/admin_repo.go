package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type AdminRepository interface {
	CheckCategoryInBoard(boardUid uint, catUid uint) bool
	GetLowestCategoryUid(boardUid uint) uint
	InsertCategory(boardUid uint, name string) uint
	IsAddedCategory(boardUid uint, name string) bool
	UpdateBoardSetting(boardUid uint, column string, value string)
	UpdatePostCategory(boardUid uint, oldCatUid uint, newCatUid uint)
	RemoveCategory(boardUid uint, catUid uint)
}

type TsboardAdminRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardAdminRepository(db *sql.DB) *TsboardAdminRepository {
	return &TsboardAdminRepository{db: db}
}

// 카테고리가 게시판에 속해 있는 것인지 확인
func (r *TsboardAdminRepository) CheckCategoryInBoard(boardUid uint, catUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT board_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, catUid).Scan(&uid)
	return boardUid == uid
}

// 가장 낮은 카테고리 고유 번호값 가져오기
func (r *TsboardAdminRepository) GetLowestCategoryUid(boardUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ? ORDER BY uid ASC LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, boardUid).Scan(&uid)
	return uid
}

// 카테고리 추가하기
func (r *TsboardAdminRepository) InsertCategory(boardUid uint, name string) uint {
	query := fmt.Sprintf("INSERT INTO %s%s (board_uid, name) VALUES (?, ?)", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	result, err := r.db.Exec(query, boardUid, name)
	if err != nil {
		return models.FAILED
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 이미 동일한 이름의 카테고리가 있는지 검사하기
func (r *TsboardAdminRepository) IsAddedCategory(boardUid uint, name string) bool {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ? AND name = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, boardUid, name).Scan(&uid)
	return uid > 0
}

// 게시판 설정 업데이트하는 쿼리 실행
func (r *TsboardAdminRepository) UpdateBoardSetting(boardUid uint, column string, value string) {
	query := fmt.Sprintf("UPDATE %s%s SET %s = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD, column)
	r.db.Exec(query, value, boardUid)
}

// 카테고리 삭제 후 게시글들의 카테고리 번호를 기본값으로 변경하기
func (r *TsboardAdminRepository) UpdatePostCategory(boardUid uint, oldCatUid uint, newCatUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET category_uid = ? WHERE board_uid = ? AND category_uid = ?",
		configs.Env.Prefix, models.TABLE_POST)
	r.db.Exec(query, newCatUid, boardUid, oldCatUid)
}

// 카테고리 삭제하기
func (r *TsboardAdminRepository) RemoveCategory(boardUid uint, catUid uint) {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.Exec(query, catUid)
}

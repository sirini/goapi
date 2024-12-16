package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type AdminRepository interface {
	CheckCategoryInBoard(boardUid uint, catUid uint) bool
	FindBoardIdTypeByUid(boardUid uint) (string, models.Board)
	FindWriterByUid(userUid uint) models.BoardWriter
	GetAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetGroupBoardList(table models.Table, bunch uint) []models.Pair
	GetLatestPosts(bunch uint) []models.AdminDashboardLatestContent
	GetLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error)
	GetLowestCategoryUid(boardUid uint) uint
	GetMemberList(bunch uint) []models.BoardWriter
	GetPointPolicy(boardUid uint) (models.BoardActionPoint, error)
	InsertCategory(boardUid uint, name string) uint
	IsAddedCategory(boardUid uint, name string) bool
	UpdateBoardAdmin(boardUid uint, newAdminUid uint) error
	UpdateBoardSetting(boardUid uint, column string, value string)
	UpdateLevelPolicy(boardUid uint, level models.BoardActionLevel) error
	UpdatePointPolicy(boardUid uint, point models.BoardActionPoint) error
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

// 게시판 아이디와 타입 반환하기
func (r *TsboardAdminRepository) FindBoardIdTypeByUid(boardUid uint) (string, models.Board) {
	var id string
	var boardType models.Board
	query := fmt.Sprintf("SELECT id, type FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&id, &boardType)
	return id, boardType
}

// 게시글 작성자 기본 정보 반환하기
func (r *TsboardAdminRepository) FindWriterByUid(userUid uint) models.BoardWriter {
	result := models.BoardWriter{}
	query := fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)

	result.UserUid = userUid
	r.db.QueryRow(query, userUid).Scan(&result.Name, &result.Profile, &result.Signature)
	return result
}

// 게시판 관리자 후보 목록 가져오기 (이름으로 검색)
func (r *TsboardAdminRepository) GetAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error) {
	items := []models.BoardWriter{}
	query := fmt.Sprintf("SELECT uid, name, profile, signature FROM %s%s WHERE blocked = ? AND name LIKE ? LIMIT ?",
		configs.Env.Prefix, models.TABLE_USER)

	rows, err := r.db.Query(query, 0, "%"+name+"%", bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardWriter{}
		err = rows.Scan(&item.UserUid, &item.Name, &item.Profile, &item.Signature)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// 대시보드용 그룹 or 게시판 목록 가져오기
func (r *TsboardAdminRepository) GetGroupBoardList(table models.Table, bunch uint) []models.Pair {
	items := []models.Pair{}
	query := fmt.Sprintf("SELECT uid, id FROM %s%s ORDER BY uid DESC LIMIT ?", configs.Env.Prefix, table)
	rows, err := r.db.Query(query, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Pair{}
		err = rows.Scan(&item.Uid, &item.Name)
		if err != nil {
			return items
		}
		items = append(items, item)
	}
	return items
}

// 대시보드용 최근 게시글 목록 가져오기
func (r *TsboardAdminRepository) GetLatestPosts(bunch uint) []models.AdminDashboardLatestContent {
	items := []models.AdminDashboardLatestContent{}
	query := fmt.Sprintf("SELECT uid, board_uid, user_uid, title FROM %s%s ORDER BY uid DESC LIMIT ?",
		configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminDashboardLatestContent{}
		var boardUid, userUid uint
		err = rows.Scan(&item.Uid, &boardUid, &userUid, &item.Content)
		if err != nil {
			return items
		}

		boardId, boardType := r.FindBoardIdTypeByUid(item.Uid)
		item.Writer = r.FindWriterByUid(userUid)
		item.Id = boardId
		item.Type = boardType
		items = append(items, item)
	}
	return items
}

// 게시판 레벨 제한값 가져오기
func (r *TsboardAdminRepository) GetLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error) {
	result := models.AdminBoardLevelPolicy{}
	result.Uid = boardUid
	query := fmt.Sprintf(`SELECT admin_uid, level_list, level_view, level_write, level_comment, level_download 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	err := r.db.QueryRow(query, boardUid).Scan(
		&result.Admin.UserUid,
		&result.Level.List,
		&result.Level.View,
		&result.Level.Write,
		&result.Level.Comment,
		&result.Level.Download,
	)
	if err != nil {
		return result, err
	}
	return result, nil
}

// 가장 낮은 카테고리 고유 번호값 가져오기
func (r *TsboardAdminRepository) GetLowestCategoryUid(boardUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ? ORDER BY uid ASC LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, boardUid).Scan(&uid)
	return uid
}

// 대시보드용 회원 목록 가져오기
func (r *TsboardAdminRepository) GetMemberList(bunch uint) []models.BoardWriter {
	items := []models.BoardWriter{}
	query := fmt.Sprintf("SELECT uid, name, profile, signature FROM %s%s ORDER BY uid DESC LIMIT ?",
		configs.Env.Prefix, models.TABLE_USER)
	rows, err := r.db.Query(query, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardWriter{}
		err = rows.Scan(&item.UserUid, &item.Name, &item.Profile, &item.Signature)
		if err != nil {
			return items
		}
		items = append(items, item)
	}
	return items
}

// 게시판 포인트 정책 가져오기
func (r *TsboardAdminRepository) GetPointPolicy(boardUid uint) (models.BoardActionPoint, error) {
	result := models.BoardActionPoint{}
	query := fmt.Sprintf(`SELECT point_view, point_write, point_comment, point_download 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	err := r.db.QueryRow(query, boardUid).Scan(&result.View, &result.Write, &result.Comment, &result.Download)
	if err != nil {
		return result, err
	}
	return result, nil
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

// 게시판 관리자 변경하기
func (r *TsboardAdminRepository) UpdateBoardAdmin(boardUid uint, newAdminUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET admin_uid = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, newAdminUid, boardUid)
	return err
}

// 게시판 레벨 제한 변경하기
func (r *TsboardAdminRepository) UpdateLevelPolicy(boardUid uint, level models.BoardActionLevel) error {
	query := fmt.Sprintf(`UPDATE %s%s SET level_list = ?, level_view = ?, level_write = ?, level_comment = ?, level_download = ? 
												WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, level.List, level.View, level.Write, level.Comment, level.Download)
	return err
}

// 게시판 포인트 정책 변경하기
func (r *TsboardAdminRepository) UpdatePointPolicy(boardUid uint, point models.BoardActionPoint) error {
	query := fmt.Sprintf(`UPDATE %s%s SET point_view = ?, point_write = ?, point_comment = ?, point_download = ? 
												WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, point.View, point.Write, point.Comment, point.Download, boardUid)
	return err
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

package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type AdminRepository interface {
	CheckCategoryInBoard(boardUid uint, catUid uint) bool
	CreateBoard(groupUid uint, newBoardId string) uint
	CreateDefaultCategories(boardUid uint, cats []string)
	CreateGroup(newGroupId string) uint
	FindPathByUid(table models.Table, targetUid uint) []string
	FindBoardIdTypeByUid(boardUid uint) (string, models.Board)
	FindBoardUidByPostUid(postUid uint) uint
	FindBoardInfoById(inputId string, bunch uint) []models.Triple
	FindGroupUidAdminUidById(groupId string) (uint, uint)
	FindGroupUidIdById(inputId string, bunch uint) []models.Pair
	FindLikeByUid(table models.Table, targetUid uint) uint
	FindThumbPathByPostUid(postUid uint) []string
	FindWriterByUid(userUid uint) models.BoardWriter
	GetAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetCommentCount(postUid uint) uint
	GetComments(rows *sql.Rows) []models.AdminLatestComment
	GetDashboardComments(bunch uint) []models.AdminDashboardLatestContent
	GetDashboardPosts(bunch uint) []models.AdminDashboardLatestContent
	GetDashboardReports(bunch uint) []models.AdminDashboardReport
	GetDefaultGroupUid(exceptUid uint) uint
	GetGroupBoardList(table models.Table, bunch uint) []models.Pair
	GetGroupList() []models.AdminGroupItem
	GetLatestComments(page uint, bunch uint, maxUid uint) []models.AdminLatestComment
	GetLatestPosts(page uint, bunch uint, maxUid uint) []models.AdminLatestPost
	GetLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error)
	GetLowestCategoryUid(boardUid uint) uint
	GetReportList(param models.AdminReportParameter) []models.AdminReportItem
	GetMemberList(bunch uint) []models.BoardWriter
	GetPointPolicy(boardUid uint) (models.BoardActionPoint, error)
	GetPosts(rows *sql.Rows) []models.AdminLatestPost
	GetRemoveFilePaths(boardUid uint) []string
	GetStatistic(table models.Table, column models.StatisticColumn, days int) models.AdminDashboardStatistic
	GetSearchedComments(param models.AdminLatestParameter) []models.AdminLatestComment
	GetSearchedPosts(param models.AdminLatestParameter) []models.AdminLatestPost
	GetSearchedReports(param models.AdminReportParameter) []models.AdminReportItem
	GetTotalBoardCount(groupUid uint) uint
	GetTotalGroupCount() uint
	InsertCategory(boardUid uint, name string) uint
	IsAddedCategory(boardUid uint, name string) bool
	IsAdded(table models.Table, boardId string) bool
	UpdateBoardSetting(boardUid uint, column string, value string)
	UpdateGroupBoardAdmin(table models.Table, targetUid uint, newAdminUid uint) error
	UpdateGroupId(groupUid uint, newGroupId string) error
	UpdateGroupUid(newGroupUid uint, oldGroupUid uint) error
	UpdateLevelPolicy(boardUid uint, level models.BoardActionLevel) error
	UpdatePointPolicy(boardUid uint, point models.BoardActionPoint) error
	UpdatePostCategory(boardUid uint, oldCatUid uint, newCatUid uint)
	UpdateStatusRemoved(table models.Table, boardUid uint) error
	RemoveBoardCategories(boardUid uint) error
	RemoveBoard(boardUid uint) error
	RemoveCategory(boardUid uint, catUid uint)
	RemoveGroup(groupUid uint) error
	RemoveFileRecords(boardUid uint) error
	RemoveRecordByFileUid(table models.Table, fileUid uint)
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

// 새 게시판 만들기
func (r *TsboardAdminRepository) CreateBoard(groupUid uint, newBoardId string) uint {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(id, group_uid, admin_uid, type, name, info,
													row_count, width, use_category, level_list, level_view, level_write,
													level_comment, level_download, point_view, point_write, point_comment, point_download) 
													VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		configs.Env.Prefix, models.TABLE_BOARD)
	result, err := r.db.Exec(
		query,
		newBoardId,
		groupUid,
		models.CREATE_BOARD_ADMIN,
		models.CREATE_BOARD_TYPE,
		models.CREATE_BOARD_NAME,
		models.CREATE_BOARD_INFO,
		models.CREATE_BOARD_ROWS,
		models.CREATE_BOARD_WIDTH,
		models.CREATE_BOARD_USE_CAT,
		models.CREATE_BOARD_LV_LIST,
		models.CREATE_BOARD_LV_VIEW,
		models.CREATE_BOARD_LV_WRITE,
		models.CREATE_BOARD_LV_COMMENT,
		models.CREATE_BOARD_LV_DOWNLOAD,
		models.CREATE_BOARD_PT_VIEW,
		models.CREATE_BOARD_PT_WRITE,
		models.CREATE_BOARD_PT_COMMENT,
		models.CREATE_BOARD_PT_DOWNLOAD,
	)
	if err != nil {
		return models.FAILED
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 새 게시판 생성 시 함께 생성되는 기본 분류들 생성 처리
func (r *TsboardAdminRepository) CreateDefaultCategories(boardUid uint, cats []string) {
	for _, cat := range cats {
		query := fmt.Sprintf("INSERT INTO %s%s (board_uid, name) VALUES (?, ?)", configs.Env.Prefix, models.TABLE_BOARD_CAT)
		r.db.Exec(query, boardUid, cat)
	}
}

// 새 그룹 생성하기
func (r *TsboardAdminRepository) CreateGroup(newGroupId string) uint {
	query := fmt.Sprintf("INSERT INTO %s%s (id, admin_uid, timestamp) VALUES (?, ?, ?)", configs.Env.Prefix, models.TABLE_GROUP)
	result, err := r.db.Exec(query, newGroupId, models.CREATE_GROUP_ADMIN, time.Now().UnixMilli())
	if err != nil {
		return models.FAILED
	}

	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 게시판 삭제 시 게시글에 딸린 첨부파일들 or 본문에 삽입한 이미지들 삭제를 위한 경로 반환
func (r *TsboardAdminRepository) FindPathByUid(table models.Table, targetUid uint) []string {
	var paths []string
	query := fmt.Sprintf("SELECT path FROM %s%s WHERE %s_uid = ?", configs.Env.Prefix, table, table)
	rows, err := r.db.Query(query, targetUid)
	if err != nil {
		return paths
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		err = rows.Scan(&path)
		if err != nil {
			return paths
		}
		paths = append(paths, path)
	}
	return paths
}

// 게시판 아이디와 타입 반환하기
func (r *TsboardAdminRepository) FindBoardIdTypeByUid(boardUid uint) (string, models.Board) {
	var id string
	var boardType models.Board
	query := fmt.Sprintf("SELECT id, type FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&id, &boardType)
	return id, boardType
}

// 게시글 번호로 게시판 고유 번호 가져오기
func (r *TsboardAdminRepository) FindBoardUidByPostUid(postUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT board_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, postUid).Scan(&uid)
	return uid
}

// 입력된 게시판 아이디와 유사한 것들 가져오기
func (r *TsboardAdminRepository) FindBoardInfoById(inputId string, bunch uint) []models.Triple {
	items := []models.Triple{}
	query := fmt.Sprintf("SELECT uid, id, name FROM %s%s WHERE id LIKE ? LIMIT ?", configs.Env.Prefix, models.TABLE_BOARD)
	rows, err := r.db.Query(query, "%"+inputId+"%", bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Triple{}
		err = rows.Scan(&item.Uid, &item.Id, &item.Name)
		if err != nil {
			return items
		}
		items = append(items, item)
	}
	return items
}

// 그룹 아이디에 해당하는 고유 번호와 관리자 고유 번호 가져오기
func (r *TsboardAdminRepository) FindGroupUidAdminUidById(groupId string) (uint, uint) {
	var groupUid, adminUid uint
	query := fmt.Sprintf("SELECT uid, admin_uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	r.db.QueryRow(query, groupId).Scan(&groupUid, &adminUid)
	return groupUid, adminUid
}

// 입력된 그룹 ID가 이미 등록되었는지 확인하기 위해 유사 ID 목록 가져오기
func (r *TsboardAdminRepository) FindGroupUidIdById(inputId string, bunch uint) []models.Pair {
	items := []models.Pair{}
	query := fmt.Sprintf("SELECT uid, id FROM %s%s WHERE id LIKE ? LIMIT ?", configs.Env.Prefix, models.TABLE_GROUP)
	rows, err := r.db.Query(query, "%"+inputId+"%", bunch)
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

// 게시판 or 댓글의 좋아요 갯수 가져오기
func (r *TsboardAdminRepository) FindLikeByUid(table models.Table, targetUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE %s_uid = ? AND liked = ?", configs.Env.Prefix, table, table)
	r.db.QueryRow(query, targetUid, 1).Scan(&count)
	return count
}

// 게시판 삭제 시 게시글에 딸린 썸네일들 삭제를 위한 경로 반환
func (r *TsboardAdminRepository) FindThumbPathByPostUid(postUid uint) []string {
	var paths []string
	query := fmt.Sprintf("SELECT path, full_path FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_FILE_THUMB)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return paths
	}
	defer rows.Close()

	for rows.Next() {
		var thumb, full string
		err = rows.Scan(&thumb, &full)
		if err != nil {
			return paths
		}
		paths = append(paths, thumb, full)
	}
	return paths
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

// 게시글에 달린 댓글 갯수 가져오기
func (r *TsboardAdminRepository) GetCommentCount(postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.QueryRow(query, postUid).Scan(&count)
	return count
}

// (검색 or 일반) 댓글 정제해서 반환
func (r *TsboardAdminRepository) GetComments(rows *sql.Rows) []models.AdminLatestComment {
	items := []models.AdminLatestComment{}
	for rows.Next() {
		item := models.AdminLatestComment{}
		err := rows.Scan(&item.Uid, &item.PostUid, &item.Writer.UserUid, &item.Content, &item.Date, &item.Status)
		if err != nil {
			return items
		}
		item.Writer = r.FindWriterByUid(item.Writer.UserUid)
		item.Like = r.FindLikeByUid(models.TABLE_COMMENT, item.Uid)
		boardUid := r.FindBoardUidByPostUid(item.PostUid)
		boardId, boardType := r.FindBoardIdTypeByUid(boardUid)
		item.Id = boardId
		item.Type = boardType
		items = append(items, item)
	}
	return items
}

// 기본 그룹 번호 가져오기
func (r *TsboardAdminRepository) GetDefaultGroupUid(exceptUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE uid != ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	r.db.QueryRow(query, exceptUid).Scan(&uid)
	return uid
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

// 그룹 목록 가져오기
func (r *TsboardAdminRepository) GetGroupList() []models.AdminGroupItem {
	items := []models.AdminGroupItem{}
	query := fmt.Sprintf("SELECT uid, id, admin_uid FROM %s%s", configs.Env.Prefix, models.TABLE_GROUP)
	rows, err := r.db.Query(query)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminGroupItem{}
		err = rows.Scan(&item.Uid, &item.Id, &item.Manager.UserUid)
		if err != nil {
			return items
		}
		item.Manager = r.FindWriterByUid(item.Manager.UserUid)
		item.Count = r.GetTotalBoardCount(item.Uid)
		items = append(items, item)
	}
	return items
}

// 대시보드용 최근 댓글 목록 가져오기
func (r *TsboardAdminRepository) GetDashboardComments(bunch uint) []models.AdminDashboardLatestContent {
	items := []models.AdminDashboardLatestContent{}
	query := fmt.Sprintf("SELECT uid, post_uid, user_uid, content FROM %s%s ORDER BY uid DESC LIMIT ?",
		configs.Env.Prefix, models.TABLE_COMMENT)
	rows, err := r.db.Query(query, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminDashboardLatestContent{}
		var postUid, userUid uint
		err = rows.Scan(&item.Uid, &postUid, &userUid, &item.Content)
		if err != nil {
			return items
		}
		boardUid := r.FindBoardUidByPostUid(postUid)
		boardId, boardType := r.FindBoardIdTypeByUid(boardUid)
		item.Writer = r.FindWriterByUid(userUid)
		item.Id = boardId
		item.Type = boardType
		items = append(items, item)
	}
	return items
}

// 대시보드용 최근 게시글 목록 가져오기
func (r *TsboardAdminRepository) GetDashboardPosts(bunch uint) []models.AdminDashboardLatestContent {
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

		boardId, boardType := r.FindBoardIdTypeByUid(boardUid)
		item.Writer = r.FindWriterByUid(userUid)
		item.Id = boardId
		item.Type = boardType
		items = append(items, item)
	}
	return items
}

// 대시보드용 최근 신고 목록 가져오기
func (r *TsboardAdminRepository) GetDashboardReports(bunch uint) []models.AdminDashboardReport {
	items := []models.AdminDashboardReport{}
	query := fmt.Sprintf("SELECT uid, from_uid, request FROM %s%s ORDER BY uid DESC LIMIT ?",
		configs.Env.Prefix, models.TABLE_REPORT)
	rows, err := r.db.Query(query, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminDashboardReport{}
		err = rows.Scan(&item.Uid, &item.Writer.UserUid, &item.Content)
		if err != nil {
			return items
		}
		item.Writer = r.FindWriterByUid(item.Writer.UserUid)
		items = append(items, item)
	}
	return items
}

// 최근 댓글 가져오기
func (r *TsboardAdminRepository) GetLatestComments(page uint, bunch uint, maxUid uint) []models.AdminLatestComment {
	items := []models.AdminLatestComment{}
	last := 1 + maxUid - (page-1)*bunch
	query := fmt.Sprintf(`SELECT uid, post_uid, user_uid, content, submitted, status FROM %s%s 
												WHERE uid < ? ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_COMMENT)
	rows, err := r.db.Query(query, last, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()
	return r.GetComments(rows)
}

// 최근 게시글 가져오기
func (r *TsboardAdminRepository) GetLatestPosts(page uint, bunch uint, maxUid uint) []models.AdminLatestPost {
	items := []models.AdminLatestPost{}
	last := 1 + maxUid - (page-1)*bunch
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, title, submitted, hit, status 
												FROM %s%s WHERE uid < ? ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, last, bunch)
	if err != nil {
		return items
	}
	defer rows.Close()
	return r.GetPosts(rows)
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

// 신고 목록 가져오기
func (r *TsboardAdminRepository) GetReportList(param models.AdminReportParameter) []models.AdminReportItem {
	//
	//
	// TODO
	//
	//
	return nil
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

// (검색 or 일반) 게시글 정제해서 반환
func (r *TsboardAdminRepository) GetPosts(rows *sql.Rows) []models.AdminLatestPost {
	items := []models.AdminLatestPost{}
	for rows.Next() {
		item := models.AdminLatestPost{}
		var boardUid uint
		err := rows.Scan(&item.Uid, &boardUid, &item.Writer.UserUid, &item.Title, &item.Date, &item.Hit, &item.Status)
		if err != nil {
			return items
		}
		item.Comment = r.GetCommentCount(item.Uid)
		item.Writer = r.FindWriterByUid(item.Writer.UserUid)
		item.Like = r.FindLikeByUid(models.TABLE_POST, item.Uid)
		boardId, boardType := r.FindBoardIdTypeByUid(boardUid)
		item.Id = boardId
		item.Type = boardType
		items = append(items, item)
	}
	return items
}

// 게시판 삭제 시 제거 필요한 파일 목록 반환하기
func (r *TsboardAdminRepository) GetRemoveFilePaths(boardUid uint) []string {
	paths := []string{}
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, boardUid)
	if err != nil {
		return paths
	}
	defer rows.Close()

	for rows.Next() {
		var postUid uint
		err = rows.Scan(&postUid)
		if err != nil {
			return paths
		}

		attaches := r.FindPathByUid(models.TABLE_FILE, postUid)
		thumbs := r.FindThumbPathByPostUid(postUid)
		paths = append(paths, attaches...)
		paths = append(paths, thumbs...)
	}

	inserted := r.FindPathByUid(models.TABLE_IMAGE, boardUid)
	paths = append(paths, inserted...)
	return paths
}

// 대시보드용 각종 통계 데이터 반환
func (r *TsboardAdminRepository) GetStatistic(table models.Table, column models.StatisticColumn, days int) models.AdminDashboardStatistic {
	result := models.AdminDashboardStatistic{}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", configs.Env.Prefix, table)
	err := r.db.QueryRow(query).Scan(&result.Total)
	if err != nil {
		return result
	}

	now := time.Now()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	columnName := column.String()

	for d := 0; d < days; d++ {
		start := day.AddDate(0, 0, d*-1).UnixMilli()
		end := day.AddDate(0, 0, (d+1)*-1).UnixMilli()

		history := models.AdminDashboardStatus{}
		history.Date = uint64(start)
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE %s BETWEEN ? AND ?", configs.Env.Prefix, table, columnName)
		err = r.db.QueryRow(query, end, start).Scan(&history.Visit)
		if err != nil {
			return result
		}
		result.History = append(result.History, history)
	}
	return result
}

// 댓글 내용 검색하기
func (r *TsboardAdminRepository) GetSearchedComments(param models.AdminLatestParameter) []models.AdminLatestComment {
	items := []models.AdminLatestComment{}
	last := 1 + param.MaxUid - (param.Page-1)*param.Bunch
	query := fmt.Sprintf(`SELECT uid, post_uid, user_uid, content, submitted, status FROM %s%s 
												WHERE uid < ? AND %s LIKE ? ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_COMMENT, param.Option.String())
	rows, err := r.db.Query(query, last, "%"+param.Keyword+"%", param.Bunch)
	if err != nil {
		return items
	}
	defer rows.Close()
	return r.GetComments(rows)
}

// 게시글 내용 검색하기
func (r *TsboardAdminRepository) GetSearchedPosts(param models.AdminLatestParameter) []models.AdminLatestPost {
	items := []models.AdminLatestPost{}
	last := 1 + param.MaxUid - (param.Page-1)*param.Bunch
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, title, submitted, hit, status 
												FROM %s%s WHERE uid < ? AND %s LIKE ? ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, param.Option.String())
	rows, err := r.db.Query(query, last, "%"+param.Keyword+"%", param.Bunch)
	if err != nil {
		return items
	}
	defer rows.Close()
	return r.GetPosts(rows)
}

// 신고 목록 검색하기
func (r *TsboardAdminRepository) GetSearchedReports(param models.AdminReportParameter) []models.AdminReportItem {
	//
	//
	// TODO
	//
	//
	return nil
}

// 지정된 그룹에 소속된 게시판 갯수 반환
func (r *TsboardAdminRepository) GetTotalBoardCount(groupUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE group_uid = ?", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, groupUid).Scan(&count)
	return count
}

// 그룹의 총 갯수 반환
func (r *TsboardAdminRepository) GetTotalGroupCount() uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", configs.Env.Prefix, models.TABLE_GROUP)
	r.db.QueryRow(query).Scan(&count)
	return count
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

// 이미 추가된 그룹 or 게시판 ID인지 검사하기
func (r *TsboardAdminRepository) IsAdded(table models.Table, boardId string) bool {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, table)
	r.db.QueryRow(query, boardId).Scan(&count)
	return count > 0
}

// 게시판 설정 업데이트하는 쿼리 실행
func (r *TsboardAdminRepository) UpdateBoardSetting(boardUid uint, column string, value string) {
	query := fmt.Sprintf("UPDATE %s%s SET %s = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD, column)
	r.db.Exec(query, value, boardUid)
}

// 그룹 or 게시판 관리자 변경하기
func (r *TsboardAdminRepository) UpdateGroupBoardAdmin(table models.Table, targetUid uint, newAdminUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET admin_uid = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, table)
	_, err := r.db.Exec(query, newAdminUid, targetUid)
	return err
}

// 그룹 ID 변경하기
func (r *TsboardAdminRepository) UpdateGroupId(groupUid uint, newGroupId string) error {
	query := fmt.Sprintf("UPDATE %s%s SET id = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	_, err := r.db.Exec(query, newGroupId, groupUid)
	return err
}

// 소속 그룹 번호를 일괄 변경하기
func (r *TsboardAdminRepository) UpdateGroupUid(newGroupUid uint, oldGroupUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET group_uid = ? WHERE group_uid = ?", configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, newGroupUid, oldGroupUid)
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

// 게시판 삭제 시 (댓)글의 상태를 삭제됨으로 변경
func (r *TsboardAdminRepository) UpdateStatusRemoved(table models.Table, boardUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE board_uid = ?", configs.Env.Prefix, table)
	_, err := r.db.Exec(query, models.CONTENT_REMOVED, boardUid)
	return err
}

// 게시판 삭제 시 게시판에 속한 분류명들 삭제하기
func (r *TsboardAdminRepository) RemoveBoardCategories(boardUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	_, err := r.db.Exec(query, boardUid)
	return err
}

// 게시판 삭제하기
func (r *TsboardAdminRepository) RemoveBoard(boardUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, boardUid)
	return err
}

// 카테고리 삭제하기
func (r *TsboardAdminRepository) RemoveCategory(boardUid uint, catUid uint) {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.Exec(query, catUid)
}

// 그룹 삭제하기
func (r *TsboardAdminRepository) RemoveGroup(groupUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	_, err := r.db.Exec(query, groupUid)
	return err
}

// 게시판 삭제 시 파일 경로들 삭제하기 (주의: 실제 파일들 삭제 처리 이후 실행 필요)
func (r *TsboardAdminRepository) RemoveFileRecords(boardUid uint) error {
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_FILE)
	rows, err := r.db.Query(query, boardUid)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var fileUid uint
		err = rows.Scan(&fileUid)
		if err != nil {
			return err
		}

		r.RemoveRecordByFileUid(models.TABLE_FILE_THUMB, fileUid)
		r.RemoveRecordByFileUid(models.TABLE_EXIF, fileUid)
		r.RemoveRecordByFileUid(models.TABLE_IMAGE_DESC, fileUid)
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_FILE)
	_, err = r.db.Exec(query, boardUid)
	return err
}

// 게시판 삭제 시 레코드 삭제 필요한 테이블 작업 처리
func (r *TsboardAdminRepository) RemoveRecordByFileUid(table models.Table, fileUid uint) {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE file_uid = ?", configs.Env.Prefix, table)
	r.db.Exec(query, fileUid)
}

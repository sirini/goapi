package repositories

import (
	"database/sql"
	"fmt"
	"strings"
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
	FindCountByBoardUid(table models.Table, boardUid uint) uint
	FindWriterByUid(userUid uint) models.BoardWriter
	FindWriterUidByName(name string) uint
	GetAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error)
	GetBoardList(groupUid uint) []models.AdminGroupBoardItem
	GetCommentCount(postUid uint) uint
	GetCommentList(param models.AdminLatestParam) []models.AdminLatestComment
	GetDefaultGroupUid(exceptUid uint) uint
	GetGroupBoardList(table models.Table, bunch uint) []models.Pair
	GetGroupList() []models.AdminGroupConfig
	GetLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error)
	GetLowestCategoryUid(boardUid uint) uint
	GetReportList(param models.AdminReportParam) []models.AdminReportItem
	GetMemberList(bunch uint) []models.BoardWriter
	GetPointPolicy(boardUid uint) (models.BoardActionPoint, error)
	GetPostList(param models.AdminLatestParam) []models.AdminLatestPost
	GetRemoveFilePaths(boardUid uint) []string
	GetStatistic(table models.Table, column models.StatisticColumn, days int) models.AdminDashboardStatistic
	GetTotalBoardCount(groupUid uint) uint
	GetTotalCount(table models.Table) uint
	GetUserList(param models.AdminUserParam) []models.AdminUserItem
	GetUserInfo(userUid uint) models.AdminUserInfo
	InsertCategory(boardUid uint, name string) uint
	IsAddedCategory(boardUid uint, name string) bool
	IsAdded(table models.Table, boardId string) bool
	UpdateBoardSetting(boardUid uint, column string, value string) error
	UpdateGroupBoardAdmin(table models.Table, targetUid uint, newAdminUid uint) error
	UpdateGroupId(groupUid uint, newGroupId string) error
	UpdateGroupUid(newGroupUid uint, oldGroupUid uint) error
	UpdateLevelPolicy(boardUid uint, level models.BoardActionLevel) error
	UpdatePointPolicy(boardUid uint, point models.BoardActionPoint) error
	UpdatePostCategory(boardUid uint, oldCatUid uint, newCatUid uint) error
	UpdateStatusRemoved(table models.Table, boardUid uint) error
	UpdateUserLevelPoint(userUid uint, level uint, point uint) error
	RemoveBoardCategories(boardUid uint) error
	RemoveBoard(boardUid uint) error
	RemoveCategory(boardUid uint, catUid uint) error
	RemoveGroup(groupUid uint) error
	RemoveFileRecords(boardUid uint) error
	RemoveRecordByFileUid(table models.Table, fileUid uint) error
}

type NuboAdminRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboAdminRepository(db *sql.DB) *NuboAdminRepository {
	return &NuboAdminRepository{db: db}
}

// 카테고리가 게시판에 속해 있는 것인지 확인
func (r *NuboAdminRepository) CheckCategoryInBoard(boardUid uint, catUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT board_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, catUid).Scan(&uid)
	return boardUid == uid
}

// 새 게시판 만들기
func (r *NuboAdminRepository) CreateBoard(groupUid uint, newBoardId string) uint {
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
func (r *NuboAdminRepository) CreateDefaultCategories(boardUid uint, cats []string) {
	for _, cat := range cats {
		query := fmt.Sprintf("INSERT INTO %s%s (board_uid, name) VALUES (?, ?)", configs.Env.Prefix, models.TABLE_BOARD_CAT)
		r.db.Exec(query, boardUid, cat)
	}
}

// 새 그룹 생성하기
func (r *NuboAdminRepository) CreateGroup(newGroupId string) uint {
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
func (r *NuboAdminRepository) FindPathByUid(table models.Table, targetUid uint) []string {
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
func (r *NuboAdminRepository) FindBoardIdTypeByUid(boardUid uint) (string, models.Board) {
	var id string
	var boardType models.Board
	query := fmt.Sprintf("SELECT id, type FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&id, &boardType)
	return id, boardType
}

// 게시글 번호로 게시판 고유 번호 가져오기
func (r *NuboAdminRepository) FindBoardUidByPostUid(postUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT board_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, postUid).Scan(&uid)
	return uid
}

// 입력된 게시판 아이디와 유사한 것들 가져오기
func (r *NuboAdminRepository) FindBoardInfoById(inputId string, bunch uint) []models.Triple {
	items := make([]models.Triple, 0)
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
func (r *NuboAdminRepository) FindGroupUidAdminUidById(groupId string) (uint, uint) {
	var groupUid, adminUid uint
	query := fmt.Sprintf("SELECT uid, admin_uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	r.db.QueryRow(query, groupId).Scan(&groupUid, &adminUid)
	return groupUid, adminUid
}

// 입력된 그룹 ID가 이미 등록되었는지 확인하기 위해 유사 ID 목록 가져오기
func (r *NuboAdminRepository) FindGroupUidIdById(inputId string, bunch uint) []models.Pair {
	items := make([]models.Pair, 0)
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

// 게시판 or 댓글의 좋아요 개수 가져오기
func (r *NuboAdminRepository) FindLikeByUid(table models.Table, targetUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE %s_uid = ? AND liked = ?", configs.Env.Prefix, table, table)
	r.db.QueryRow(query, targetUid, 1).Scan(&count)
	return count
}

// 게시판 삭제 시 게시글에 딸린 썸네일들 삭제를 위한 경로 반환
func (r *NuboAdminRepository) FindThumbPathByPostUid(postUid uint) []string {
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

// 게시판 번호에 해당하는 총 레코드 수 반환
func (r *NuboAdminRepository) FindCountByBoardUid(table models.Table, boardUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, table)
	r.db.QueryRow(query, boardUid).Scan(&count)
	return count
}

// 게시글 작성자 기본 정보 반환하기
func (r *NuboAdminRepository) FindWriterByUid(userUid uint) models.BoardWriter {
	result := models.BoardWriter{}
	query := fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)

	result.UserUid = userUid
	r.db.QueryRow(query, userUid).Scan(&result.Name, &result.Profile, &result.Signature)
	return result
}

// 사용자 이름으로 고유 번화 반환하기
func (r *NuboAdminRepository) FindWriterUidByName(name string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name LIKE ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)
	r.db.QueryRow(query, "%"+name+"%").Scan(&uid)
	return uid
}

// 게시판 관리자 후보 목록 가져오기 (이름으로 검색)
func (r *NuboAdminRepository) GetAdminCandidates(name string, bunch uint) ([]models.BoardWriter, error) {
	items := make([]models.BoardWriter, 0)
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

// 그룹 소속 게시판의 기본 정보 및 간단 통계 가져오기
func (r *NuboAdminRepository) GetBoardList(groupUid uint) []models.AdminGroupBoardItem {
	items := make([]models.AdminGroupBoardItem, 0)
	query := fmt.Sprintf("SELECT uid, id, admin_uid, type, name, info FROM %s%s WHERE group_uid = ?",
		configs.Env.Prefix, models.TABLE_BOARD)

	rows, err := r.db.Query(query, groupUid)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminGroupBoardItem{}
		err = rows.Scan(&item.Uid, &item.Id, &item.Manager.UserUid, &item.Type, &item.Name, &item.Info)
		if err != nil {
			return items
		}
		item.Manager = r.FindWriterByUid(item.Manager.UserUid)
		item.Total.Post = r.FindCountByBoardUid(models.TABLE_POST, item.Uid)
		item.Total.Comment = r.FindCountByBoardUid(models.TABLE_COMMENT, item.Uid)
		item.Total.File = r.FindCountByBoardUid(models.TABLE_FILE, item.Uid)
		item.Total.Image = r.FindCountByBoardUid(models.TABLE_IMAGE, item.Uid)
		items = append(items, item)
	}
	return items
}

// 게시글에 달린 댓글 개수 가져오기
func (r *NuboAdminRepository) GetCommentCount(postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.QueryRow(query, postUid).Scan(&count)
	return count
}

// (검색된) 댓글 목록 가져오기
func (r *NuboAdminRepository) GetCommentList(param models.AdminLatestParam) []models.AdminLatestComment {
	items := make([]models.AdminLatestComment, 0)
	whereClauses := []string{"1=1"}
	whereArgs := []any{}

	if len(param.Keyword) > 0 {
		keyword := "%" + param.Keyword + "%"
		switch param.Option {
		case models.SEARCH_WRITER:
			whereClauses = append(whereClauses, "t_user.name LIKE ?")
		case models.SEARCH_CONTENT:
			whereClauses = append(whereClauses, "c.content LIKE ?")
		}
		whereArgs = append(whereArgs, keyword)
	}

	whereQuery := strings.Join(whereClauses, " AND ")
	offset := (param.Page - 1) * param.Limit
	query := fmt.Sprintf(`SELECT c.uid, c.post_uid, c.user_uid, c.content, c.submitted, c.status,
			t_user.name, t_user.profile,
			t_board.id, t_board.type, t_board.name,
			COALESCE(l.like_count, 0)
		FROM %s%s c
		JOIN (
			SELECT c.uid FROM %s%s c
			LEFT JOIN %s%s AS t_user ON c.user_uid = t_user.uid
			WHERE %s 
			ORDER BY c.uid DESC 
			LIMIT ? OFFSET ?
		) AS sub ON c.uid = sub.uid
		LEFT JOIN %s%s AS t_user ON c.user_uid = t_user.uid
		LEFT JOIN %s%s AS t_board ON c.board_uid = t_board.uid
		LEFT JOIN (
			SELECT comment_uid, COUNT(*) AS like_count FROM %s%s WHERE liked = ?
			GROUP BY comment_uid
		) AS l ON c.uid = l.comment_uid
		ORDER BY c.uid DESC`,
		configs.Env.Prefix, models.TABLE_COMMENT,
		configs.Env.Prefix, models.TABLE_COMMENT,
		configs.Env.Prefix, models.TABLE_USER,
		whereQuery,
		configs.Env.Prefix, models.TABLE_USER,
		configs.Env.Prefix, models.TABLE_BOARD,
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE,
	)

	whereArgs = append(whereArgs, param.Limit, offset, 1)
	rows, err := r.db.Query(query, whereArgs...)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminLatestComment{}
		err := rows.Scan(&item.Uid, &item.PostUid, &item.Writer.UserUid, &item.Content, &item.Date, &item.Status,
			&item.Writer.Name, &item.Writer.Profile,
			&item.Id, &item.Type, &item.Name, &item.Like)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

// 기본 그룹 번호 가져오기
func (r *NuboAdminRepository) GetDefaultGroupUid(exceptUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE uid != ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	r.db.QueryRow(query, exceptUid).Scan(&uid)
	return uid
}

// 대시보드용 그룹 or 게시판 목록 가져오기
func (r *NuboAdminRepository) GetGroupBoardList(table models.Table, bunch uint) []models.Pair {
	items := make([]models.Pair, 0)
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
func (r *NuboAdminRepository) GetGroupList() []models.AdminGroupConfig {
	items := make([]models.AdminGroupConfig, 0)
	query := fmt.Sprintf("SELECT uid, id, admin_uid FROM %s%s", configs.Env.Prefix, models.TABLE_GROUP)
	rows, err := r.db.Query(query)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminGroupConfig{}
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

// 게시판 레벨 제한값 가져오기
func (r *NuboAdminRepository) GetLevelPolicy(boardUid uint) (models.AdminBoardLevelPolicy, error) {
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
func (r *NuboAdminRepository) GetLowestCategoryUid(boardUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ? ORDER BY uid ASC LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, boardUid).Scan(&uid)
	return uid
}

// 신고 목록 가져오기
func (r *NuboAdminRepository) GetReportList(param models.AdminReportParam) []models.AdminReportItem {
	items := make([]models.AdminReportItem, 0)

	isSolved := 0
	if param.IsSolved {
		isSolved = 1
	}

	whereClauses := []string{"solved = ?"}
	whereArgs := []any{isSolved}

	if len(param.Keyword) > 0 {
		switch param.Option {
		case models.SEARCH_REPORT_TO, models.SEARCH_REPORT_FROM:
			col := "t_user.name"
			if param.Option == models.SEARCH_REPORT_FROM {
				col = "f_user.name"
			}
			whereClauses = append(whereClauses, fmt.Sprintf("%s LIKE ?", col))
			whereArgs = append(whereArgs, "%"+param.Keyword+"%")
		case models.SEARCH_REPORT_REQUEST:
			whereClauses = append(whereClauses, "request LIKE ?")
			whereArgs = append(whereArgs, "%"+param.Keyword+"%")
		}
	}

	whereQuery := strings.Join(whereClauses, " AND ")
	offset := (param.Page - 1) * param.Limit
	query := fmt.Sprintf(`SELECT r.uid,
			r.to_uid, r.from_uid, r.request, r.response, r.timestamp,
			t_user.name, f_user.name
		FROM %s%s r
		JOIN (
			SELECT uid FROM %s%s WHERE %s ORDER BY uid DESC LIMIT ? OFFSET ?
		) AS sub ON r.uid = sub.uid
		LEFT JOIN %s%s AS t_user ON r.to_uid = t_user.uid
		LEFT JOIN %s%s AS f_user ON r.from_uid = f_user.uid
		ORDER BY r.uid DESC`,
		configs.Env.Prefix, models.TABLE_REPORT,
		configs.Env.Prefix, models.TABLE_REPORT,
		whereQuery,
		configs.Env.Prefix, models.TABLE_USER,
		configs.Env.Prefix, models.TABLE_USER,
	)

	whereArgs = append(whereArgs, param.Limit, offset)
	rows, err := r.db.Query(query, whereArgs...)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminReportItem{}
		err := rows.Scan(
			&item.Uid, &item.To.UserUid, &item.From.UserUid,
			&item.Request, &item.Response, &item.Date,
			&item.To.Name, &item.From.Name,
		)
		if err == nil {
			items = append(items, item)
		}
	}
	return items
}

// 대시보드용 회원 목록 가져오기
func (r *NuboAdminRepository) GetMemberList(bunch uint) []models.BoardWriter {
	items := make([]models.BoardWriter, 0)
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
func (r *NuboAdminRepository) GetPointPolicy(boardUid uint) (models.BoardActionPoint, error) {
	result := models.BoardActionPoint{}
	query := fmt.Sprintf(`SELECT point_view, point_write, point_comment, point_download 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	err := r.db.QueryRow(query, boardUid).Scan(&result.View, &result.Write, &result.Comment, &result.Download)
	if err != nil {
		return result, err
	}
	return result, nil
}

// (검색된) 게시글 가져오기
func (r *NuboAdminRepository) GetPostList(param models.AdminLatestParam) []models.AdminLatestPost {
	items := make([]models.AdminLatestPost, 0)
	whereClauses := []string{"1=1"}
	whereArgs := []any{}

	if len(param.Keyword) > 0 {
		keyword := "%" + param.Keyword + "%"
		switch param.Option {
		case models.SEARCH_TITLE:
			whereClauses = append(whereClauses, "p.title LIKE ?")
		case models.SEARCH_CONTENT:
			whereClauses = append(whereClauses, "p.content LIKE ?")
		case models.SEARCH_WRITER:
			whereClauses = append(whereClauses, "t_user.name LIKE ?")
		}
		whereArgs = append(whereArgs, keyword)
	}

	whereQuery := strings.Join(whereClauses, " AND ")
	offset := (param.Page - 1) * param.Limit
	query := fmt.Sprintf(`SELECT p.uid, p.user_uid, p.title, p.submitted, p.hit, p.status,
			t_user.name, t_user.profile,
			t_board.id, t_board.type, t_board.name,
			COALESCE(c.comment_count, 0),
			COALESCE(l.like_count, 0)
		FROM %s%s p
		JOIN (
			SELECT p.uid FROM %s%s p
			LEFT JOIN %s%s AS t_user ON p.user_uid = t_user.uid
			WHERE %s
			ORDER BY p.uid DESC
			LIMIT ? OFFSET ?
		)	AS sub ON p.uid = sub.uid
		LEFT JOIN %s%s AS t_user ON p.user_uid = t_user.uid
		LEFT JOIN %s%s AS t_board ON p.board_uid = t_board.uid
		LEFT JOIN (
			SELECT post_uid, COUNT(*) AS comment_count FROM %s%s
			GROUP BY post_uid
		) AS c ON p.uid = c.post_uid
		LEFT JOIN (
			SELECT post_uid, COUNT(*) AS like_count FROM %s%s WHERE liked = ?
			GROUP BY post_uid
		) AS l ON p.uid = l.post_uid
		ORDER BY p.uid DESC`,
		configs.Env.Prefix, models.TABLE_POST,
		configs.Env.Prefix, models.TABLE_POST,
		configs.Env.Prefix, models.TABLE_USER,
		whereQuery,
		configs.Env.Prefix, models.TABLE_USER,
		configs.Env.Prefix, models.TABLE_BOARD,
		configs.Env.Prefix, models.TABLE_COMMENT,
		configs.Env.Prefix, models.TABLE_POST_LIKE,
	)

	whereArgs = append(whereArgs, param.Limit, offset, 1)
	rows, err := r.db.Query(query, whereArgs...)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminLatestPost{}
		err := rows.Scan(&item.Uid, &item.Writer.UserUid, &item.Title, &item.Date, &item.Hit, &item.Status,
			&item.Writer.Name, &item.Writer.Profile, &item.Id, &item.Type, &item.Name, &item.Comment, &item.Like)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

// 게시판 삭제 시 제거 필요한 파일 목록 반환하기
func (r *NuboAdminRepository) GetRemoveFilePaths(boardUid uint) []string {
	paths := make([]string, 0)
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
func (r *NuboAdminRepository) GetStatistic(table models.Table, column models.StatisticColumn, days int) models.AdminDashboardStatistic {
	result := models.AdminDashboardStatistic{
		History: make([]models.AdminDashboardStatus, 0, days),
	}
	prefix := configs.Env.Prefix
	columnName := column.String()
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", prefix, table)
	_ = r.db.QueryRow(totalQuery).Scan(&result.Total)

	historyQuery := fmt.Sprintf(`SELECT DATE_FORMAT(FROM_UNIXTIME(%s / 1000), '%%Y-%%m-%%d') AS date_str, COUNT(*) AS cnt
		FROM %s%s WHERE %s >= UNIX_TIMESTAMP(DATE_SUB(CURDATE(), INTERVAL ? DAY)) * 1000
		GROUP BY date_str ORDER BY date_str DESC`, columnName, prefix, table, columnName)

	rows, err := r.db.Query(historyQuery, days-1)
	if err != nil {
		return result
	}
	defer rows.Close()

	statsMap := make(map[string]uint64)
	for rows.Next() {
		var dateStr string
		var count uint64
		if err := rows.Scan(&dateStr, &count); err == nil {
			statsMap[dateStr] = count
		}
	}

	now := time.Now()
	for d := range days {
		targetDate := now.AddDate(0, 0, -d)
		dateKey := targetDate.Format("2006-01-02")
		history := models.AdminDashboardStatus{
			Date:  uint64(targetDate.UnixMilli()),
			Visit: uint(statsMap[dateKey]),
		}
		result.History = append(result.History, history)
	}
	return result
}

// 지정된 그룹에 소속된 게시판 개수 반환
func (r *NuboAdminRepository) GetTotalBoardCount(groupUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE group_uid = ?", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, groupUid).Scan(&count)
	return count
}

// 지정 테이블의 총 레코드 개수 반환
func (r *NuboAdminRepository) GetTotalCount(table models.Table) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", configs.Env.Prefix, table)
	r.db.QueryRow(query).Scan(&count)
	return count
}

// (검색된) 사용자 목록 반환
func (r *NuboAdminRepository) GetUserList(param models.AdminUserParam) []models.AdminUserItem {
	items := make([]models.AdminUserItem, 0)
	last := (param.Page - 1) * param.Limit
	isBlockedQuery := "<"
	if param.IsBlocked {
		isBlockedQuery = "="
	}

	whereQuery := ""
	switch param.Option {
	case models.SEARCH_USER_NAME:
		whereQuery = "AND name LIKE '%" + param.Keyword + "%'"
	case models.SEARCH_USER_ID:
		whereQuery = "AND id LIKE '%" + param.Keyword + "%'"
	case models.SEARCH_USER_LEVEL:
		whereQuery = "AND level = " + param.Keyword
	}

	query := fmt.Sprintf(`SELECT uid, id, name, profile, level, point, signup 
												FROM %s%s WHERE uid < ? AND blocked %s 1 %s
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_USER, isBlockedQuery, whereQuery)
	rows, err := r.db.Query(query, last, param.Limit)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.AdminUserItem{}
		err = rows.Scan(&item.UserUid, &item.Id, &item.Name, &item.Profile, &item.Level, &item.Point, &item.Signup)
		if err != nil {
			return items
		}
		items = append(items, item)
	}
	return items
}

// 사용자 정보 반환
func (r *NuboAdminRepository) GetUserInfo(userUid uint) models.AdminUserInfo {
	result := models.AdminUserInfo{}
	query := fmt.Sprintf("SELECT uid, id, name, profile, level, point, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	r.db.QueryRow(query, userUid).Scan(&result.UserUid, &result.Id, &result.Name, &result.Profile, &result.Level, &result.Point, &result.Signature)
	return result
}

// 카테고리 추가하기
func (r *NuboAdminRepository) InsertCategory(boardUid uint, name string) uint {
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
func (r *NuboAdminRepository) IsAddedCategory(boardUid uint, name string) bool {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE board_uid = ? AND name = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, boardUid, name).Scan(&uid)
	return uid > 0
}

// 이미 추가된 그룹 or 게시판 ID인지 검사하기
func (r *NuboAdminRepository) IsAdded(table models.Table, boardId string) bool {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, table)
	r.db.QueryRow(query, boardId).Scan(&count)
	return count > 0
}

// 게시판 설정 업데이트하는 쿼리 실행
func (r *NuboAdminRepository) UpdateBoardSetting(boardUid uint, column string, value string) error {
	query := fmt.Sprintf("UPDATE %s%s SET %s = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD, column)
	_, err := r.db.Exec(query, value, boardUid)
	return err
}

// 그룹 or 게시판 관리자 변경하기
func (r *NuboAdminRepository) UpdateGroupBoardAdmin(table models.Table, targetUid uint, newAdminUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET admin_uid = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, table)
	_, err := r.db.Exec(query, newAdminUid, targetUid)
	return err
}

// 그룹 ID 변경하기
func (r *NuboAdminRepository) UpdateGroupId(groupUid uint, newGroupId string) error {
	query := fmt.Sprintf("UPDATE %s%s SET id = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	_, err := r.db.Exec(query, newGroupId, groupUid)
	return err
}

// 소속 그룹 번호를 일괄 변경하기
func (r *NuboAdminRepository) UpdateGroupUid(newGroupUid uint, oldGroupUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET group_uid = ? WHERE group_uid = ?", configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, newGroupUid, oldGroupUid)
	return err
}

// 게시판 레벨 제한 변경하기
func (r *NuboAdminRepository) UpdateLevelPolicy(boardUid uint, level models.BoardActionLevel) error {
	query := fmt.Sprintf(`UPDATE %s%s SET level_list = ?, level_view = ?, level_write = ?, level_comment = ?, level_download = ? 
												WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, level.List, level.View, level.Write, level.Comment, level.Download, boardUid)
	return err
}

// 게시판 포인트 정책 변경하기
func (r *NuboAdminRepository) UpdatePointPolicy(boardUid uint, point models.BoardActionPoint) error {
	query := fmt.Sprintf(`UPDATE %s%s SET point_view = ?, point_write = ?, point_comment = ?, point_download = ? 
												WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, point.View, point.Write, point.Comment, point.Download, boardUid)
	return err
}

// 카테고리 삭제 후 게시글들의 카테고리 번호를 기본값으로 변경하기
func (r *NuboAdminRepository) UpdatePostCategory(boardUid uint, oldCatUid uint, newCatUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET category_uid = ? WHERE board_uid = ? AND category_uid = ?",
		configs.Env.Prefix, models.TABLE_POST)
	_, err := r.db.Exec(query, newCatUid, boardUid, oldCatUid)
	return err
}

// 게시판 삭제 시 (댓)글의 상태를 삭제됨으로 변경
func (r *NuboAdminRepository) UpdateStatusRemoved(table models.Table, boardUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE board_uid = ?", configs.Env.Prefix, table)
	_, err := r.db.Exec(query, models.CONTENT_REMOVED, boardUid)
	return err
}

// 사용자의 레벨, 포인트 정보 변경하기
func (r *NuboAdminRepository) UpdateUserLevelPoint(userUid uint, level uint, point uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET level = ?, point = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER)
	_, err := r.db.Exec(query, level, point, userUid)
	return err
}

// 게시판 삭제 시 게시판에 속한 분류명들 삭제하기
func (r *NuboAdminRepository) RemoveBoardCategories(boardUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	_, err := r.db.Exec(query, boardUid)
	return err
}

// 게시판 삭제하기
func (r *NuboAdminRepository) RemoveBoard(boardUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	_, err := r.db.Exec(query, boardUid)
	return err
}

// 카테고리 삭제하기
func (r *NuboAdminRepository) RemoveCategory(boardUid uint, catUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	_, err := r.db.Exec(query, catUid)
	return err
}

// 그룹 삭제하기
func (r *NuboAdminRepository) RemoveGroup(groupUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_GROUP)
	_, err := r.db.Exec(query, groupUid)
	return err
}

// 게시판 삭제 시 파일 경로들 삭제하기 (주의: 실제 파일들 삭제 처리 이후 실행 필요)
func (r *NuboAdminRepository) RemoveFileRecords(boardUid uint) error {
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

		err = r.RemoveRecordByFileUid(models.TABLE_FILE_THUMB, fileUid)
		if err != nil {
			return err
		}
		err = r.RemoveRecordByFileUid(models.TABLE_EXIF, fileUid)
		if err != nil {
			return err
		}
		err = r.RemoveRecordByFileUid(models.TABLE_IMAGE_DESC, fileUid)
		if err != nil {
			return err
		}
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_FILE)
	_, err = r.db.Exec(query, boardUid)
	return err
}

// 게시판 삭제 시 레코드 삭제 필요한 테이블 작업 처리
func (r *NuboAdminRepository) RemoveRecordByFileUid(table models.Table, fileUid uint) error {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE file_uid = ?", configs.Env.Prefix, table)
	_, err := r.db.Exec(query, fileUid)
	return err
}

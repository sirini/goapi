package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardRepository interface {
	CheckLikedPostForLoop(stmt *sql.Stmt, postUid uint, userUid uint) bool
	CheckLikedPost(postUid uint, userUid uint) bool
	CheckLikedCommentForLoop(stmt *sql.Stmt, commentUid uint, userUid uint) bool
	CheckLikedComment(commentUid uint, userUid uint) bool
	FindPostsByImageDescription(param models.BoardListParam) ([]models.BoardListItem, error)
	FindPostsByTitleContent(param models.BoardListParam) ([]models.BoardListItem, error)
	FindPostsByNameCategory(param models.BoardListParam) ([]models.BoardListItem, error)
	FindPostsByHashtag(param models.BoardListParam) ([]models.BoardListItem, error)
	GetBoardConfig(boardUid uint) models.BoardConfig
	GetBoardUidById(id string) uint
	GetBoardCategories(boardUid uint) []models.Pair
	GetCategoryByUidForLoop(stmt *sql.Stmt, categoryUid uint) models.Pair
	GetCategoryByUid(categoryUid uint) models.Pair
	GetCoverImageForLoop(stmt *sql.Stmt, postUid uint) string
	GetCoverImage(postUid uint) string
	GetCommentCountForLoop(stmt *sql.Stmt, postUid uint) uint
	GetCommentCount(postUid uint) uint
	GetCommentLikeCount(postUid uint) uint
	GetGroupAdminUid(boardUid uint) uint
	GetLikeCount(postUid uint) uint
	GetLikedCountForLoop(stmt *sql.Stmt, postUid uint) uint
	GetNoticePosts(boardUid uint, actionUserUid uint) ([]models.BoardListItem, error)
	GetNormalPosts(param models.BoardListParam) ([]models.BoardListItem, error)
	GetMaxUid(table models.Table) uint
	GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error)
	GetTagUids(names string) (string, int)
	GetTotalPostCount(boardUid uint) uint
	GetUidByTable(table models.Table, name string) uint
	GetWriterInfoForLoop(stmt *sql.Stmt, userUid uint) models.BoardWriter
	GetWriterInfo(userUid uint) models.BoardWriter
	MakeListItem(actionUserUid uint, rows *sql.Rows) ([]models.BoardListItem, error)
}

type NuboBoardRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboBoardRepository(db *sql.DB) *NuboBoardRepository {
	return &NuboBoardRepository{db: db}
}

// 게시글 가져오기 시 지정되는 컬럼들
const POST_COLUMNS = "uid, user_uid, category_uid, title, content, submitted, modified, hit, status"

// 반복문 내에서 사용할 게시글에 좋아요 클릭 여부 확인하기
func (r *NuboBoardRepository) CheckLikedPostForLoop(stmt *sql.Stmt, postUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	var liked uint8
	stmt.QueryRow(postUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 게시글에 좋아요를 클릭했었는지 확인하기
func (r *NuboBoardRepository) CheckLikedPost(postUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	query := fmt.Sprintf("SELECT liked FROM %s%s WHERE post_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST_LIKE)

	var liked uint8
	r.db.QueryRow(query, postUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 반복문에서 사용하는 댓글에 좋아요 클릭했는지 확인
func (r *NuboBoardRepository) CheckLikedCommentForLoop(stmt *sql.Stmt, commentUid uint, userUid uint) bool {
	var liked uint8
	stmt.QueryRow(commentUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 댓글에 좋아요를 클릭했었는지 확인하기
func (r *NuboBoardRepository) CheckLikedComment(commentUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	query := fmt.Sprintf("SELECT liked FROM %s%s WHERE comment_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	var liked uint8
	r.db.QueryRow(query, commentUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 게시글에 첨부된 이미지에 대한 AI 분석 내용으로 검색해서 가져오기
func (r *NuboBoardRepository) FindPostsByImageDescription(param models.BoardListParam) ([]models.BoardListItem, error) {
	option := param.Option.String()
	keyword := "%" + param.Keyword + "%"
	normalLimit := param.Limit - param.NoticeCount
	offset := (param.Page - 1) * normalLimit
	query := fmt.Sprintf(`SELECT p.uid, p.user_uid, p.category_uid, p.title, p.content, p.submitted, p.modified, p.hit, p.status 
												FROM %s%s AS p
												JOIN (
													SELECT DISTINCT d.post_uid
													FROM %s%s AS d
													JOIN %s%s AS p2 ON d.post_uid = p2.uid
													WHERE p2.board_uid = ?
														AND p2.status = ?
														AND d.%s LIKE ?
													ORDER BY d.post_uid DESC
													LIMIT ? OFFSET ?
												) AS filtered ON p.uid = filtered.post_uid
												ORDER BY p.uid DESC`,
		configs.Env.Prefix, models.TABLE_POST,
		configs.Env.Prefix, models.TABLE_IMAGE_DESC,
		configs.Env.Prefix, models.TABLE_POST,
		option)
	rows, err := r.db.Query(query, param.BoardUid, models.CONTENT_NORMAL, keyword, normalLimit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시글 제목 혹은 내용으로 검색해서 가져오기
func (r *NuboBoardRepository) FindPostsByTitleContent(param models.BoardListParam) ([]models.BoardListItem, error) {
	option := param.Option.String()
	keyword := "%" + param.Keyword + "%"
	normalLimit := param.Limit - param.NoticeCount
	offset := (param.Page - 1) * normalLimit
	query := fmt.Sprintf(`SELECT t.uid, t.user_uid, t.category_uid, t.title, t.content, t.submitted, t.modified, t.hit, t.status 
												FROM %s%s AS t
												JOIN (SELECT uid FROM %s%s 
													WHERE board_uid = ? AND status = ? AND %s LIKE ?
													ORDER BY uid DESC LIMIT ? OFFSET ?) AS p
												ON t.uid = p.uid`, configs.Env.Prefix, models.TABLE_POST, configs.Env.Prefix, models.TABLE_POST, option)
	rows, err := r.db.Query(query, param.BoardUid, models.CONTENT_NORMAL, keyword, normalLimit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시글 작성자 혹은 분류명으로 검색해서 가져오기
func (r *NuboBoardRepository) FindPostsByNameCategory(param models.BoardListParam) ([]models.BoardListItem, error) {
	option := param.Option.String()
	table := models.TABLE_USER
	if param.Option == models.SEARCH_CATEGORY {
		table = models.TABLE_BOARD_CAT
	}
	uid := r.GetUidByTable(table, param.Keyword)
	normalLimit := param.Limit - param.NoticeCount
	offset := (param.Page - 1) * normalLimit
	query := fmt.Sprintf(`SELECT t.uid, t.user_uid, t.category_uid, t.title, t.content, t.submitted, t.modified, t.hit, t.status 
												FROM %s%s AS t
												JOIN (SELECT uid FROM %s%s
													WHERE board_uid = ? AND status = ? AND %s = ?
													ORDER BY uid DESC LIMIT ? OFFSET ?) AS p
												ON t.uid = p.uid`, configs.Env.Prefix, models.TABLE_POST, configs.Env.Prefix, models.TABLE_POST, option)
	rows, err := r.db.Query(query, param.BoardUid, models.CONTENT_NORMAL, uid, normalLimit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시글 태그로 검색해서 가져오기
func (r *NuboBoardRepository) FindPostsByHashtag(param models.BoardListParam) ([]models.BoardListItem, error) {
	tagUidStr, _ := r.GetTagUids(param.Keyword)
	normalLimit := param.Limit - param.NoticeCount
	offset := (param.Page - 1) * normalLimit
	query := fmt.Sprintf(`SELECT p.uid, p.user_uid, p.category_uid, p.title, p.content, p.submitted, p.modified, p.hit, p.status 
												FROM %s%s AS p
												JOIN (
													SELECT DISTINCT ph.post_uid
													FROM %s%s AS ph
													JOIN %s%s AS p2 ON ph.post_uid = p2.uid
													WHERE ph.board_uid = ?
														AND p2.status = ?
														AND ph.hashtag_uid IN (%s)
													ORDER BY ph.post_uid DESC
													LIMIT ? OFFSET ?
												) AS filtered ON p.uid = filtered.post_uid
												 ORDER BY p.uid DESC`, configs.Env.Prefix, models.TABLE_POST,
		configs.Env.Prefix, models.TABLE_POST_HASHTAG,
		configs.Env.Prefix, models.TABLE_POST, tagUidStr)
	rows, err := r.db.Query(query, param.BoardUid, models.CONTENT_NORMAL, normalLimit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시판 설정값 가져오기
func (r *NuboBoardRepository) GetBoardConfig(boardUid uint) models.BoardConfig {
	config := models.BoardConfig{}
	query := fmt.Sprintf(`SELECT id, group_uid, admin_uid, type, name, info, row_count, width, use_category,
												level_list, level_view, level_write, level_comment, level_download,
												point_view, point_write, point_comment, point_download 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)

	var useCategory uint8
	r.db.QueryRow(query, boardUid).Scan(&config.Id, &config.GroupUid, &config.Admin.Board, &config.Type, &config.Name, &config.Info,
		&config.RowCount, &config.Width, &useCategory, &config.Level.List, &config.Level.View,
		&config.Level.Write, &config.Level.Comment, &config.Level.Download, &config.Point.View,
		&config.Point.Write, &config.Point.Comment, &config.Point.Download)
	config.Uid = boardUid
	config.UseCategory = useCategory > 0
	config.Category = r.GetBoardCategories(boardUid)
	config.Admin.Group = r.GetGroupAdminUid(boardUid)
	return config
}

// 게시판 아이디로 게시판 고유 번호 가져오기
func (r *NuboBoardRepository) GetBoardUidById(id string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)

	r.db.QueryRow(query, id).Scan(&uid)
	return uid
}

// 지정된 게시판에서 사용중인 카테고리 목록들 반환
func (r *NuboBoardRepository) GetBoardCategories(boardUid uint) []models.Pair {
	items := make([]models.Pair, 0)
	query := fmt.Sprintf("SELECT uid, name FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_BOARD_CAT)

	rows, err := r.db.Query(query, boardUid)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Pair{}
		err := rows.Scan(&item.Uid, &item.Name)
		if err != nil {
			return items
		}
		items = append(items, item)
	}
	return items
}

// 반복문에서 사용할 카테고리 이름 가져오기
func (r *NuboBoardRepository) GetCategoryByUidForLoop(stmt *sql.Stmt, categoryUid uint) models.Pair {
	cat := models.Pair{}
	stmt.QueryRow(categoryUid).Scan(&cat.Uid, &cat.Name)
	return cat
}

// 카테고리 이름 가져오기
func (r *NuboBoardRepository) GetCategoryByUid(categoryUid uint) models.Pair {
	cat := models.Pair{}
	query := fmt.Sprintf("SELECT uid, name FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)

	r.db.QueryRow(query, categoryUid).Scan(&cat.Uid, &cat.Name)
	return cat
}

// 게시글 대표 커버 썸네일 이미지 가져오기
func (r *NuboBoardRepository) GetCoverImageForLoop(stmt *sql.Stmt, postUid uint) string {
	var path string
	stmt.QueryRow(postUid).Scan(&path)
	return path
}

// 반복문에서 사용하는 썸네일 이미지 가져오기
func (r *NuboBoardRepository) GetCoverImage(postUid uint) string {
	query := fmt.Sprintf("SELECT path FROM %s%s WHERE post_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)

	var path string
	r.db.QueryRow(query, postUid).Scan(&path)
	return path
}

// 댓글 개수 가져오기
func (r *NuboBoardRepository) GetCommentCount(postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE post_uid = ? AND status != ?", configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, postUid, models.CONTENT_REMOVED).Scan(&count)
	return count
}

// 댓글에 대한 좋아요 개수 가져오기
func (r *NuboBoardRepository) GetCommentLikeCount(postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE post_uid = ? AND liked = ?", configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.QueryRow(query, postUid, 1).Scan(&count)
	return count
}

// 반복문에서 사용하는 댓글 개수 가져오기
func (r *NuboBoardRepository) GetCommentCountForLoop(stmt *sql.Stmt, postUid uint) uint {
	var count uint
	stmt.QueryRow(postUid, models.CONTENT_REMOVED).Scan(&count)
	return count
}

// 게시판이 속한 그룹의 관리자 고유 번호값 가져오기
func (r *NuboBoardRepository) GetGroupAdminUid(boardUid uint) uint {
	var adminUid uint
	query := fmt.Sprintf(`SELECT g.admin_uid FROM %s%s AS g JOIN %s%s AS b 
												ON g.uid = b.group_uid WHERE b.uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_GROUP, configs.Env.Prefix, models.TABLE_BOARD)

	r.db.QueryRow(query, boardUid).Scan(&adminUid)
	return adminUid
}

// 좋아요 개수 가져오기
func (r *NuboBoardRepository) GetLikeCount(postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE post_uid = ? AND liked = ?", configs.Env.Prefix, models.TABLE_POST_LIKE)

	r.db.QueryRow(query, postUid, 1).Scan(&count)
	return count
}

// 반복문에서 사용하는 게시글의 좋아요 갯수 가져오기
func (r *NuboBoardRepository) GetLikedCountForLoop(stmt *sql.Stmt, postUid uint) uint {
	var count uint
	stmt.QueryRow(postUid, 1).Scan(&count)
	return count
}

// 게시판 공지글만 가져오기
func (r *NuboBoardRepository) GetNoticePosts(boardUid uint, actionUserUid uint) ([]models.BoardListItem, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s%s WHERE board_uid = ? AND status = ?`,
		POST_COLUMNS, configs.Env.Prefix, models.TABLE_POST)

	rows, err := r.db.Query(query, boardUid, models.CONTENT_NOTICE)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(actionUserUid, rows)
}

// 비밀글을 포함한 일반 게시글들 가져오기
func (r *NuboBoardRepository) GetNormalPosts(param models.BoardListParam) ([]models.BoardListItem, error) {
	normalLimit := param.Limit - param.NoticeCount
	offset := (param.Page - 1) * normalLimit
	query := fmt.Sprintf(`SELECT t.uid, t.user_uid, t.category_uid, t.title, t.content, t.submitted, t.modified, t.hit, t.status 
												FROM %s%s AS t
												JOIN (SELECT uid FROM %s%s 
													WHERE board_uid = ? AND status IN (?, ?) 
													ORDER BY uid DESC LIMIT ? OFFSET ?) AS p
												ON t.uid = p.uid`, configs.Env.Prefix, models.TABLE_POST, configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, param.BoardUid, models.CONTENT_NORMAL, models.CONTENT_SECRET, normalLimit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 최근 사용된 해시태그 가져오기
func (r *NuboBoardRepository) GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error) {
	items := make([]models.BoardTag, 0)
	query := fmt.Sprintf(`SELECT h.uid AS hashtag_uid, h.name AS hashtag_name, IFNULL(MAX(p.post_uid), 0) AS latest_post 
												FROM %s%s p 
												JOIN %s%s h ON p.hashtag_uid = h.uid 
												WHERE board_uid = ? 
												GROUP BY h.uid, h.name 
												ORDER BY latest_post 
												DESC limit ?`,
		configs.Env.Prefix, models.TABLE_POST_HASHTAG, configs.Env.Prefix, models.TABLE_HASHTAG)
	rows, err := r.db.Query(query, boardUid, limit)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardTag{}
		err = rows.Scan(&item.Uid, &item.Name, &item.PostUid)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}

// 게시판 or 댓글의 현재 uid 값 반환하기
func (r *NuboBoardRepository) GetMaxUid(table models.Table) uint {
	var max uint
	query := fmt.Sprintf("SELECT IFNULL(MAX(uid), 0) FROM %s%s", configs.Env.Prefix, table)
	r.db.QueryRow(query).Scan(&max)
	return max
}

// 스페이스로 구분된 태그 이름들을 가져와서 태그 고유번호 문자열로 변환
func (r *NuboBoardRepository) GetTagUids(keyword string) (string, int) {
	tags := strings.Split(keyword, " ")
	var strUids []string

	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_HASHTAG)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return "", 0
	}
	defer stmt.Close()

	for _, tag := range tags {
		var uid uint
		if err := stmt.QueryRow(tag).Scan(&uid); err != nil {
			continue
		}
		strUids = append(strUids, fmt.Sprintf("'%d'", uid))
	}
	result := strings.Join(strUids, ",")
	return result, len(strUids)
}

// 게시판에 등록된 글 갯수 반환
func (r *NuboBoardRepository) GetTotalPostCount(boardUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE board_uid = ? AND status != ?",
		configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, boardUid, models.CONTENT_REMOVED).Scan(&count)
	return count
}

// 이름으로 고유 번호 가져오기 (회원 번호 혹은 카테고리 번호 등)
func (r *NuboBoardRepository) GetUidByTable(table models.Table, name string) uint {
	var uid uint
	value := "%" + name + "%"
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name LIKE ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix, table)

	r.db.QueryRow(query, value).Scan(&uid)
	return uid
}

// 반복문에서 사용하는 (댓)글 작성자 기본 정보 가져오기
func (r *NuboBoardRepository) GetWriterInfoForLoop(stmt *sql.Stmt, userUid uint) models.BoardWriter {
	writer := models.BoardWriter{}
	writer.UserUid = userUid
	stmt.QueryRow(userUid).Scan(&writer.Name, &writer.Profile, &writer.Signature)
	return writer
}

// (댓)글 작성자 기본 정보 가져오기
func (r *NuboBoardRepository) GetWriterInfo(userUid uint) models.BoardWriter {
	writer := models.BoardWriter{}
	query := fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)

	writer.UserUid = userUid
	r.db.QueryRow(query, userUid).Scan(&writer.Name, &writer.Profile, &writer.Signature)
	return writer
}

// 게시글 목록 만들어서 반환
func (r *NuboBoardRepository) MakeListItem(actionUserUid uint, rows *sql.Rows) ([]models.BoardListItem, error) {
	items := make([]models.BoardListItem, 0)

	// 카테고리 이름 가져오는 쿼리문 준비
	query := fmt.Sprintf("SELECT uid, name FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	stmtBoardCat, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtBoardCat.Close()

	// 커버 이미지 가져오는 쿼리문 준비
	query = fmt.Sprintf("SELECT path FROM %s%s WHERE post_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)
	stmtFileThumb, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtFileThumb.Close()

	// 댓글 개수 가져오는 쿼리문 준비
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE post_uid = ? AND status != ?",
		configs.Env.Prefix, models.TABLE_COMMENT)
	stmtCommmentCount, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtCommmentCount.Close()

	// 좋아요 개수 가져오는 쿼리문 준비
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE post_uid = ? AND liked = ?",
		configs.Env.Prefix, models.TABLE_POST_LIKE)
	stmtLikeCount, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtLikeCount.Close()

	// 내가 이 게시글에 좋아요를 눌렀는지 확인하는 쿼리문 준비
	query = fmt.Sprintf("SELECT liked FROM %s%s WHERE post_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST_LIKE)
	stmtLiked, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtLiked.Close()

	// 게시글 작성자 정보 가져오는 쿼리문 준비
	query = fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmtWriter, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtWriter.Close()

	for rows.Next() {
		item := models.BoardListItem{}
		var writerUid uint
		err := rows.Scan(&item.Uid, &writerUid, &item.Category.Uid, &item.Title, &item.Content,
			&item.Submitted, &item.Modified, &item.Hit, &item.Status)
		if err != nil {
			return nil, err
		}

		item.Category = r.GetCategoryByUidForLoop(stmtBoardCat, item.Category.Uid)
		item.Cover = r.GetCoverImageForLoop(stmtFileThumb, item.Uid)
		item.Comment = r.GetCommentCountForLoop(stmtCommmentCount, item.Uid)
		item.Like = r.GetLikedCountForLoop(stmtLikeCount, item.Uid)
		item.Liked = r.CheckLikedPostForLoop(stmtLiked, item.Uid, actionUserUid)
		item.Writer = r.GetWriterInfoForLoop(stmtWriter, writerUid)
		items = append(items, item)
	}
	return items, nil
}

package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardRepository interface {
	CheckLikedPost(postUid uint, userUid uint) bool
	CheckLikedComment(commentUid uint, userUid uint) bool
	GetBoardConfig(boardUid uint) models.BoardConfig
	GetBoardUidById(id string) uint
	GetBoardCategories(boardUid uint) []models.Pair
	GetCategoryByUid(categoryUid uint) models.Pair
	GetCoverImage(postUid uint) string
	GetCommentCount(postUid uint) uint
	GetCommentLikeCount(postUid uint) uint
	GetGroupAdminUid(boardUid uint) uint
	GetLikeCount(postUid uint) uint
	GetNoticePosts(boardUid uint, actionUserUid uint) ([]models.BoardListItem, error)
	GetMaxUid(table models.Table) uint
	GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error)
	GetTagUids(names string) (string, int)
	GetUidByTable(table models.Table, name string) uint
	GetWriterInfo(userUid uint) models.BoardWriter
	GetTotalCount(param models.BoardListParam) uint
	FindPosts(param models.BoardListParam) ([]models.BoardListItem, error)
}

type NuboBoardRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboBoardRepository(db *sql.DB) *NuboBoardRepository {
	return &NuboBoardRepository{db: db}
}

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

// 카테고리 이름 가져오기
func (r *NuboBoardRepository) GetCategoryByUid(categoryUid uint) models.Pair {
	cat := models.Pair{}
	query := fmt.Sprintf("SELECT uid, name FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)

	r.db.QueryRow(query, categoryUid).Scan(&cat.Uid, &cat.Name)
	return cat
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

// 이름으로 고유 번호 가져오기 (회원 번호 혹은 카테고리 번호 등)
func (r *NuboBoardRepository) GetUidByTable(table models.Table, name string) uint {
	var uid uint
	value := "%" + name + "%"
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name LIKE ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix, table)

	r.db.QueryRow(query, value).Scan(&uid)
	return uid
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

// 공지글 목록 가져오기
func (r *NuboBoardRepository) GetNoticePosts(boardUid uint, actionUserUid uint) ([]models.BoardListItem, error) {
	items := make([]models.BoardListItem, 0)
	prefix := configs.Env.Prefix

	query := fmt.Sprintf(`SELECT 
			p.uid, p.user_uid, p.category_uid, p.title, p.content, p.submitted, p.modified, p.hit, p.status,
			u.name, u.profile,
			c.name,
			COALESCE((SELECT path FROM %s%s WHERE post_uid = p.uid LIMIT 1), ''),
			(SELECT COUNT(*) FROM %s%s WHERE post_uid = p.uid AND status != ?),
			(SELECT COUNT(*) FROM %s%s WHERE post_uid = p.uid AND liked = 1),
			EXISTS(SELECT 1 FROM %s%s WHERE post_uid = p.uid AND user_uid = ? AND liked = 1)
		FROM %s%s AS p
		JOIN (
			SELECT uid FROM %s%s WHERE board_uid = ? AND status = ?
			ORDER BY uid DESC
		) AS sub ON p.uid = sub.uid
		LEFT JOIN %s%s AS u ON p.user_uid = u.uid
		LEFT JOIN %s%s AS c ON p.category_uid = c.uid`,
		prefix, models.TABLE_FILE_THUMB,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_POST_LIKE,
		prefix, models.TABLE_POST_LIKE,
		prefix, models.TABLE_POST,
		prefix, models.TABLE_POST,
		prefix, models.TABLE_USER,
		prefix, models.TABLE_BOARD_CAT,
	)

	// 파라미터 바인딩 순서 확인
	rows, err := r.db.Query(query, models.CONTENT_REMOVED, actionUserUid, boardUid, models.CONTENT_NOTICE)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardListItem{}
		err = rows.Scan(
			&item.Uid, &item.Writer.UserUid, &item.Category.Uid, &item.Title, &item.Content,
			&item.Submitted, &item.Modified, &item.Hit, &item.Status,
			&item.Writer.Name, &item.Writer.Profile,
			&item.Category.Name,
			&item.Cover,
			&item.Comment,
			&item.Like,
			&item.Liked,
		)
		if err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// 검색 조건에 따른 총 게시글 수 반환
func (r *NuboBoardRepository) GetTotalCount(param models.BoardListParam) uint {
	var count uint
	prefix := configs.Env.Prefix

	whereClauses := []string{"board_uid = ?"}
	args := []any{param.BoardUid}

	whereClauses = append(whereClauses, "status IN (?, ?)")
	args = append(args, models.CONTENT_NORMAL, models.CONTENT_SECRET)

	if len(param.Keyword) > 0 {
		switch param.Option {
		case models.SEARCH_IMAGE_DESC:
			whereClauses = append(whereClauses, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM %s%s AS d 
				WHERE d.post_uid = uid AND d.%s LIKE ?
			)`, prefix, models.TABLE_IMAGE_DESC, param.Option.String()))
			args = append(args, "%"+param.Keyword+"%")

		case models.SEARCH_TAG:
			tagUidStr, _ := r.GetTagUids(param.Keyword)
			whereClauses = append(whereClauses, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM %s%s AS ph 
				WHERE ph.post_uid = uid AND ph.hashtag_uid IN (%s)
			)`, prefix, models.TABLE_POST_HASHTAG, tagUidStr))

		case models.SEARCH_WRITER, models.SEARCH_CATEGORY:
			table := models.TABLE_USER
			if param.Option == models.SEARCH_CATEGORY {
				table = models.TABLE_BOARD_CAT
			}
			uid := r.GetUidByTable(table, param.Keyword)
			whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", param.Option.String()))
			args = append(args, uid)

		default:
			whereClauses = append(whereClauses, fmt.Sprintf("%s LIKE ?", param.Option.String()))
			args = append(args, "%"+param.Keyword+"%")
		}
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE %s",
		prefix, models.TABLE_POST, strings.Join(whereClauses, " AND "))

	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0
	}

	return count
}

// 게시글 목록 가져오기
func (r *NuboBoardRepository) FindPosts(param models.BoardListParam) ([]models.BoardListItem, error) {
	items := make([]models.BoardListItem, 0)
	normalLimit := param.Limit - param.NoticeCount
	offset := (param.Page - 1) * normalLimit

	var subQuery string
	var args []any
	prefix := configs.Env.Prefix

	if len(param.Keyword) > 0 {
		switch param.Option {
		case models.SEARCH_IMAGE_DESC:
			subQuery = fmt.Sprintf(`
            SELECT DISTINCT d.post_uid as uid FROM %s%s AS d
            JOIN %s%s AS p2 ON d.post_uid = p2.uid
            WHERE p2.board_uid = ? AND p2.status IN (?, ?) AND d.%s LIKE ?
            ORDER BY d.post_uid DESC LIMIT ? OFFSET ?`,
				prefix, models.TABLE_IMAGE_DESC,
				prefix, models.TABLE_POST,
				param.Option.String())
			args = append(args, param.BoardUid, models.CONTENT_NORMAL, models.CONTENT_SECRET, "%"+param.Keyword+"%", normalLimit, offset)

		case models.SEARCH_TAG:
			tagUids, _ := r.GetTagUids(param.Keyword)
			subQuery = fmt.Sprintf(`
            SELECT DISTINCT ph.post_uid as uid FROM %s%s AS ph
            JOIN %s%s AS p2 ON ph.post_uid = p2.uid
            WHERE ph.board_uid = ? AND p2.status IN (?, ?) AND ph.hashtag_uid IN (%s)
            ORDER BY ph.post_uid DESC LIMIT ? OFFSET ?`,
				prefix, models.TABLE_POST_HASHTAG,
				prefix, models.TABLE_POST,
				tagUids)
			args = append(args, param.BoardUid, models.CONTENT_NORMAL, models.CONTENT_SECRET, normalLimit, offset)

		case models.SEARCH_WRITER, models.SEARCH_CATEGORY:
			whereCol := param.Option.String() + " ="
			table := models.TABLE_USER
			if param.Option == models.SEARCH_CATEGORY {
				table = models.TABLE_BOARD_CAT
			}
			searchValue := r.GetUidByTable(table, param.Keyword)
			subQuery = fmt.Sprintf(`
            SELECT uid FROM %s%s 
            WHERE board_uid = ? AND status IN (?, ?) AND %s ?
            ORDER BY uid DESC LIMIT ? OFFSET ?`,
				prefix, models.TABLE_POST, whereCol)
			args = append(args, param.BoardUid, models.CONTENT_NORMAL, models.CONTENT_SECRET, searchValue, normalLimit, offset)

		case models.SEARCH_TITLE, models.SEARCH_CONTENT:
			whereCol := param.Option.String()
			searchValue := "%" + param.Keyword + "%"
			subQuery = fmt.Sprintf(`
						SELECT uid FROM %s%s
						WHERE board_uid = ? AND status IN (?, ?) AND %s LIKE ?
						ORDER BY uid DESC LIMIT ? OFFSET ?`,
				prefix, models.TABLE_POST, whereCol)
			args = append(args, param.BoardUid, models.CONTENT_NORMAL, models.CONTENT_SECRET, searchValue, normalLimit, offset)
		}
	} else {
		subQuery = fmt.Sprintf(`
            SELECT uid FROM %s%s 
            WHERE board_uid = ? AND status IN (?, ?)
            ORDER BY uid DESC LIMIT ? OFFSET ?`,
			prefix, models.TABLE_POST)
		args = append(args, param.BoardUid, models.CONTENT_NORMAL, models.CONTENT_SECRET, normalLimit, offset)
	}

	finalQuery := fmt.Sprintf(`SELECT 
            p.uid, p.user_uid, p.category_uid, p.title, p.content, p.submitted, p.modified, p.hit, p.status,
            u.name, u.profile,
            c.name,
						COALESCE((SELECT path FROM %s%s WHERE post_uid = p.uid LIMIT 1), ''),
            (SELECT COUNT(*) FROM %s%s WHERE post_uid = p.uid AND status != ?),
            (SELECT COUNT(*) FROM %s%s WHERE post_uid = p.uid AND liked = 1),
            EXISTS(SELECT 1 FROM %s%s WHERE post_uid = p.uid AND user_uid = ? AND liked = 1)
        FROM %s%s AS p
        JOIN (%s) AS sub ON p.uid = sub.uid
        LEFT JOIN %s%s AS u ON p.user_uid = u.uid
        LEFT JOIN %s%s AS c ON p.category_uid = c.uid
        ORDER BY p.uid DESC`,
		prefix, models.TABLE_FILE_THUMB,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_POST_LIKE,
		prefix, models.TABLE_POST_LIKE,
		prefix, models.TABLE_POST,
		subQuery,
		prefix, models.TABLE_USER,
		prefix, models.TABLE_BOARD_CAT,
	)

	finalArgs := []any{models.CONTENT_REMOVED}
	finalArgs = append(finalArgs, param.UserUid)
	finalArgs = append(finalArgs, args...)

	rows, err := r.db.Query(finalQuery, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardListItem{}
		err := rows.Scan(
			&item.Uid, &item.Writer.UserUid, &item.Category.Uid, &item.Title, &item.Content,
			&item.Submitted, &item.Modified, &item.Hit, &item.Status,
			&item.Writer.Name, &item.Writer.Profile,
			&item.Category.Name,
			&item.Cover,
			&item.Comment,
			&item.Like,
			&item.Liked,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

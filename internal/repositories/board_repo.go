package repositories

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardRepository interface {
	CheckLikedPost(postUid uint, userUid uint) bool
	CheckLikedComment(commentUid uint, userUid uint) bool
	FindLatestPostsByTitleContent(param *models.BoardPostParameter) ([]*models.BoardPostItem, error)
	FindLatestPostsByUserUidCatUid(param *models.BoardPostParameter) ([]*models.BoardPostItem, error)
	FindLatestPostsByTag(param *models.BoardPostParameter) ([]*models.BoardPostItem, error)
	GetBoardSettings(boardUid uint) *models.BoardBasicSettingResult
	GetBoardUidById(id string) uint
	GetCategoryNameByUid(categoryUid uint) string
	GetCountByTable(table models.Table, postUid uint) uint
	GetMaxUid() uint
	GetTagUids(names string) (string, int)
	GetUidByTable(table models.Table, name string) uint
	GetWriterInfo(userUid uint) *models.BoardWriter
	LoadLatestPosts(param *models.BoardPostParameter) ([]*models.BoardPostItem, error)
}

type TsboardBoardRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardBoardRepository(db *sql.DB) *TsboardBoardRepository {
	return &TsboardBoardRepository{db: db}
}

// 가져온 게시글 레코드들을 패킹해서 반환
func AppendItem(rows *sql.Rows) ([]*models.BoardPostItem, error) {
	var items []*models.BoardPostItem
	for rows.Next() {
		item := &models.BoardPostItem{}
		err := rows.Scan(&item.Uid, &item.BoardUid, &item.UserUid, &item.CategoryUid,
			&item.Title, &item.Content, &item.Submitted, &item.Modified, &item.Hit, &item.Status)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// 게시글에 좋아요를 클릭했었는지 확인하기
func (r *TsboardBoardRepository) CheckLikedPost(postUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	var liked uint8
	query := fmt.Sprintf("SELECT liked FROM %s%s WHERE post_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST_LIKE)
	r.db.QueryRow(query, postUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 댓글에 좋아요를 클릭했었는지 확인하기
func (r *TsboardBoardRepository) CheckLikedComment(commentUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	var liked uint8
	query := fmt.Sprintf("SELECT liked FROM %s%s WHERE comment_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	r.db.QueryRow(query, commentUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 게시글 제목 혹은 내용 일부를 검색해서 가져오기
func (r *TsboardBoardRepository) FindLatestPostsByTitleContent(param *models.BoardPostParameter) ([]*models.BoardPostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND board_uid = %d", param.BoardUid)
	}
	option := param.Option.String()
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, category_uid, 
												title, content, submitted, modified, hit, status 
												FROM %s%s WHERE status != ? %s AND %s LIKE ? 
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, whereBoard, option)
	rows, err := r.db.Query(query, models.POST_REMOVED, "%"+param.Keyword+"%", param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return AppendItem(rows)
}

// 사용자 고유 번호 혹은 게시글 카테고리 번호로 검색해서 가져오기
func (r *TsboardBoardRepository) FindLatestPostsByUserUidCatUid(param *models.BoardPostParameter) ([]*models.BoardPostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND board_uid = %d", param.BoardUid)
	}
	option := param.Option.String()
	table := models.TABLE_USER
	if param.Option == models.SEARCH_CATEGORY {
		table = models.TABLE_BOARD_CAT
	}
	uid := r.GetUidByTable(table, param.Keyword)
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, category_uid,
												title, content, submitted, modified, hit, status
												FROM %s%s WHERE status != ? %s AND %s = ?
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, whereBoard, option)
	rows, err := r.db.Query(query, models.POST_REMOVED, uid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return AppendItem(rows)
}

// 태그 이름에 해당하는 최근 게시글들만 가져오기
func (r *TsboardBoardRepository) FindLatestPostsByTag(param *models.BoardPostParameter) ([]*models.BoardPostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND p.board_uid = %d", param.BoardUid)
	}
	tagUidStr, tagCount := r.GetTagUids(param.Keyword)
	query := fmt.Sprintf(`SELECT p.uid, p.board_uid, p.user_uid, p.category_uid, 
												p.title, p.content, p.submitted, p.modified, p.hit, p.status 
												FROM %s%s AS p JOIN %s%s AS ph ON p.uid = ph.post_uid 
												WHERE p.status != ? %s AND uid < ? AND ph.hashtag_uid IN ('%s') 
												GROUP BY ph.post_uid HAVING (COUNT(ph.hashtag_uid) = ?) 
												ORDER BY p.uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, configs.Env.Prefix, models.TABLE_POST_HASHTAG, whereBoard, tagUidStr)
	rows, err := r.db.Query(query, models.POST_REMOVED, param.SinceUid, tagCount, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return AppendItem(rows)
}

// 게시판 기본 설정값 가져오기
func (r *TsboardBoardRepository) GetBoardSettings(boardUid uint) *models.BoardBasicSettingResult {
	var useCategory uint
	settings := &models.BoardBasicSettingResult{}

	query := fmt.Sprintf("SELECT id, type, use_category FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&settings.Id, &settings.Type, &useCategory)
	settings.UseCategory = useCategory > 0
	return settings
}

// 게시판 아이디로 게시판 고유 번호 가져오기
func (r *TsboardBoardRepository) GetBoardUidById(id string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, id).Scan(&uid)
	return uid
}

// 카테고리 이름 가져오기
func (r *TsboardBoardRepository) GetCategoryNameByUid(categoryUid uint) string {
	var name string
	query := fmt.Sprintf("SELECT name FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, categoryUid).Scan(&name)
	return name
}

// 댓글 혹은 좋아요 개수 가져오기
func (r *TsboardBoardRepository) GetCountByTable(table models.Table, postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, table)
	r.db.QueryRow(query, postUid).Scan(&count)
	return count
}

// 게시판의 현재 uid 값 반환하기
func (r *TsboardBoardRepository) GetMaxUid() uint {
	var max uint
	query := fmt.Sprintf("SELECT MAX(uid) AS last FROM %s%s", configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query).Scan(&max)
	return max
}

// 스페이스로 구분된 태그 이름들을 가져와서 태그 고유번호 문자열로 변환
func (r *TsboardBoardRepository) GetTagUids(keyword string) (string, int) {
	tags := strings.Split(keyword, " ")
	var strUids []string
	for _, tag := range tags {
		var uid uint
		query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1",
			configs.Env.Prefix, models.TABLE_HASHTAG)
		if err := r.db.QueryRow(query, tag).Scan(&uid); err != nil {
			continue
		}
		strUids = append(strUids, strconv.Itoa(int(uid)))
	}
	result := strings.Join(strUids, ",")
	return result, len(strUids)
}

// 이름으로 고유 번호 가져오기 (회원 번호 혹은 카테고리 번호 등)
func (r *TsboardBoardRepository) GetUidByTable(table models.Table, name string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix, table)
	r.db.QueryRow(query, name).Scan(&uid)
	return uid
}

// (댓)글 작성자 기본 정보 가져오기
func (r *TsboardBoardRepository) GetWriterInfo(userUid uint) *models.BoardWriter {
	writer := &models.BoardWriter{}
	writer.UserUid = userUid
	query := fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	r.db.QueryRow(query, userUid).Scan(&writer.Name, &writer.Profile, &writer.Signature)
	return writer
}

// 최근 게시글들 가져오기
func (r *TsboardBoardRepository) LoadLatestPosts(param *models.BoardPostParameter) ([]*models.BoardPostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND board_uid = %d", param.BoardUid)
	}
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, category_uid, 
												title, content, submitted, modified, hit, status
												FROM %s%s WHERE status != ? %s AND uid < ? 
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, whereBoard)
	rows, err := r.db.Query(query, models.POST_REMOVED, param.SinceUid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return AppendItem(rows)
}

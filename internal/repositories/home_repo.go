package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type HomeRepository interface {
	AppendItem(rows *sql.Rows) ([]models.HomePostItem, error)
	FindLatestPostsByTitleContent(param models.HomePostParameter) ([]models.HomePostItem, error)
	FindLatestPostsByUserUidCatUid(param models.HomePostParameter) ([]models.HomePostItem, error)
	FindLatestPostsByTag(param models.HomePostParameter) ([]models.HomePostItem, error)
	GetBoardBasicSettings(boardUid uint) models.BoardBasicSettingResult
	GetBoardIDs() []string
	GetBoardLinks(stmt *sql.Stmt, groupUid uint) ([]models.HomeSidebarBoardResult, error)
	GetGroupBoardLinks() ([]models.HomeSidebarGroupResult, error)
	GetLatestPosts(param models.HomePostParameter) ([]models.HomePostItem, error)
	InsertVisitorLog(userUid uint)
}

type TsboardHomeRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, boardRepo 포인터 주입받기
func NewTsboardHomeRepository(db *sql.DB, board BoardRepository) *TsboardHomeRepository {
	return &TsboardHomeRepository{
		db:    db,
		board: board,
	}
}

// 홈화면에 보여줄 게시글 레코드들을 패킹해서 반환
func (r *TsboardHomeRepository) AppendItem(rows *sql.Rows) ([]models.HomePostItem, error) {
	items := make([]models.HomePostItem, 0)
	for rows.Next() {
		item := models.HomePostItem{}
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

// 홈화면에서 게시글 제목 혹은 내용 일부를 검색해서 가져오기
func (r *TsboardHomeRepository) FindLatestPostsByTitleContent(param models.HomePostParameter) ([]models.HomePostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND board_uid = %d", param.BoardUid)
	}
	option := param.Option.String()
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, category_uid, title, content, submitted, modified, hit, status 
												FROM %s%s WHERE uid < ? AND status != ? %s AND %s LIKE ? 
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, whereBoard, option)

	rows, err := r.db.Query(query, param.SinceUid, models.CONTENT_REMOVED, "%"+param.Keyword+"%", param.Bunch)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return r.AppendItem(rows)
}

// 홈화면에서 사용자 고유 번호 혹은 게시글 카테고리 번호로 검색해서 가져오기
func (r *TsboardHomeRepository) FindLatestPostsByUserUidCatUid(param models.HomePostParameter) ([]models.HomePostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND board_uid = %d", param.BoardUid)
	}
	option := param.Option.String()
	table := models.TABLE_USER
	if param.Option == models.SEARCH_CATEGORY {
		table = models.TABLE_BOARD_CAT
	}
	uid := r.board.GetUidByTable(table, param.Keyword)
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, category_uid, title, content, submitted, modified, hit, status
												FROM %s%s WHERE uid < ? AND status != ? %s AND %s = ?
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, whereBoard, option)

	rows, err := r.db.Query(query, param.SinceUid, models.CONTENT_REMOVED, uid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.AppendItem(rows)
}

// 홈화면에서 태그 이름에 해당하는 최근 게시글들만 가져오기
func (r *TsboardHomeRepository) FindLatestPostsByTag(param models.HomePostParameter) ([]models.HomePostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND p.board_uid = %d", param.BoardUid)
	}
	tagUidStr, tagCount := r.board.GetTagUids(param.Keyword)
	query := fmt.Sprintf(`SELECT p.uid, p.board_uid, p.user_uid, p.category_uid, 
												p.title, p.content, p.submitted, p.modified, p.hit, p.status 
												FROM %s%s AS p JOIN %s%s AS ph ON p.uid = ph.post_uid 
												WHERE p.status != ? %s AND uid < ? AND ph.hashtag_uid IN (%s) 
												GROUP BY ph.post_uid HAVING (COUNT(ph.hashtag_uid) = ?) 
												ORDER BY p.uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, configs.Env.Prefix, models.TABLE_POST_HASHTAG, whereBoard, tagUidStr)

	rows, err := r.db.Query(query, models.CONTENT_REMOVED, param.SinceUid, tagCount, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.AppendItem(rows)
}

// 게시판 기본 설정값 가져오기
func (r *TsboardHomeRepository) GetBoardBasicSettings(boardUid uint) models.BoardBasicSettingResult {
	var useCategory uint
	settings := models.BoardBasicSettingResult{}

	query := fmt.Sprintf("SELECT id, type, use_category FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD)

	r.db.QueryRow(query, boardUid).Scan(&settings.Id, &settings.Type, &useCategory)
	settings.UseCategory = useCategory > 0
	return settings
}

// 전체 게시판들의 ID만 가져오기
func (r *TsboardHomeRepository) GetBoardIDs() []string {
	var result []string
	query := fmt.Sprintf("SELECT id FROM %s%s", configs.Env.Prefix, models.TABLE_BOARD)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		rows.Scan(&id)
		result = append(result, id)
	}
	return result
}

// 홈화면에서 게시판 목록들 가져오기
func (r *TsboardHomeRepository) GetBoardLinks(stmt *sql.Stmt, groupUid uint) ([]models.HomeSidebarBoardResult, error) {
	boards := make([]models.HomeSidebarBoardResult, 0)
	rows, err := stmt.Query(groupUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		board := models.HomeSidebarBoardResult{}
		if err := rows.Scan(&board.Id, &board.Type, &board.Name, &board.Info); err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return boards, nil
}

// 홈화면 사이드바에 사용할 그룹 및 하위 게시판 목록들 가져오기
func (r *TsboardHomeRepository) GetGroupBoardLinks() ([]models.HomeSidebarGroupResult, error) {
	groups := make([]models.HomeSidebarGroupResult, 0)
	query := fmt.Sprintf("SELECT uid, id FROM %s%s", configs.Env.Prefix, models.TABLE_GROUP)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 게시판 링크들을 가져오는 쿼리문 준비
	query = fmt.Sprintf("SELECT id, type, name, info FROM %s%s WHERE group_uid = ?",
		configs.Env.Prefix, models.TABLE_BOARD)
	stmtBoard, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtBoard.Close()

	for rows.Next() {
		var groupUid uint
		var groupId string
		if err := rows.Scan(&groupUid, &groupId); err != nil {
			return nil, err
		}
		boards, err := r.GetBoardLinks(stmtBoard, groupUid)
		if err != nil {
			return nil, err
		}

		group := models.HomeSidebarGroupResult{}
		group.Group = groupId
		group.Boards = boards
		groups = append(groups, group)
	}
	return groups, nil
}

// 홈화면 최근 게시글들 가져오기
func (r *TsboardHomeRepository) GetLatestPosts(param models.HomePostParameter) ([]models.HomePostItem, error) {
	whereBoard := ""
	if param.BoardUid > 0 {
		whereBoard = fmt.Sprintf("AND board_uid = %d", param.BoardUid)
	}
	query := fmt.Sprintf(`SELECT uid, board_uid, user_uid, category_uid, 
												title, content, submitted, modified, hit, status
												FROM %s%s WHERE status != ? %s AND uid < ? 
												ORDER BY uid DESC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, whereBoard)

	rows, err := r.db.Query(query, models.CONTENT_REMOVED, param.SinceUid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.AppendItem(rows)
}

// 방문자 기록하기
func (r *TsboardHomeRepository) InsertVisitorLog(userUid uint) {
	query := fmt.Sprintf("INSERT INTO %s%s (user_uid, timestamp) VALUES (?, ?)",
		configs.Env.Prefix, models.TABLE_USER_ACCESS)

	r.db.Exec(query, userUid, time.Now().UnixMilli())
}

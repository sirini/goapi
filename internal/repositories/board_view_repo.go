package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardViewRepository interface {
	CheckBannedByWriter(postUid uint, viewerUid uint) bool
	GetAllBoards() []models.BoardItem
	GetAttachments(postUid uint) ([]models.BoardAttachment, error)
	GetAttachedImages(postUid uint) ([]models.BoardAttachedImage, error)
	GetBasicBoardConfig(boardUid uint) models.BoardBasicConfig
	GetDownloadInfo(fileUid uint) models.BoardViewDownloadResult
	GetExif(fileUid uint) models.BoardExif
	GetNeededLevelPoint(boardUid uint, action models.BoardAction) (int, int)
	GetPrevPostUid(boardUid uint, postUid uint) uint
	GetNextPostUid(boardUid uint, postUid uint) uint
	GetPost(postUid uint, actionUserUid uint) (models.BoardListItem, error)
	GetTags(postUid uint) []models.Pair
	GetTagName(hashtagUid uint) string
	GetThumbnailImage(fileUid uint) models.BoardThumbnail
	GetWriterLatestComment(writerUid uint, limit uint) ([]models.BoardWriterLatestComment, error)
	GetWriterLatestPost(writerUid uint, limit uint) ([]models.BoardWriterLatestPost, error)
	InsertLikePost(param models.BoardViewLikeParameter)
	IsLikedPost(postUid uint, actionUserUid uint) bool
	IsWriter(table models.Table, postUid uint, userUid uint) bool
	RemoveAttachments(postUid uint) []string
	RemoveAttachedFile(fileUid uint, filePath string) []string
	RemovePost(postUid uint)
	RemovePostTags(postUid uint)
	RemoveThumbnails(fileUid uint) []string
	UpdateLikePost(param models.BoardViewLikeParameter)
	UpdatePostHit(postUid uint)
	UpdatePostBoardUid(targetBoardUid uint, postUid uint)
}

type TsboardBoardViewRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardBoardViewRepository(db *sql.DB, board BoardRepository) *TsboardBoardViewRepository {
	return &TsboardBoardViewRepository{db: db, board: board}
}

// 글작성자에게 차단당한 사용자인지 확인하기
func (r *TsboardBoardViewRepository) CheckBannedByWriter(postUid uint, viewerUid uint) bool {
	var writerUid uint
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, postUid).Scan(&writerUid)
	if writerUid < 1 {
		return false
	}

	var blackUid uint
	query = fmt.Sprintf("SELECT black_uid FROM %s%s WHERE user_uid = ? AND black_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_USER_BLOCK)
	r.db.QueryRow(query, writerUid, viewerUid).Scan(&blackUid)
	return blackUid > 0
}

// 게시판 목록들 가져오기 (게시글 이동 시 필요)
func (r *TsboardBoardViewRepository) GetAllBoards() []models.BoardItem {
	var items []models.BoardItem
	query := fmt.Sprintf("SELECT uid, name, info FROM %s%s", configs.Env.Prefix, models.TABLE_BOARD)
	rows, err := r.db.Query(query)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.BoardItem{}
		rows.Scan(&item.Uid, &item.Name, &item.Info)
		items = append(items, item)
	}
	return items
}

// 게시글에 첨부된 파일 목록들 가져오기
func (r *TsboardBoardViewRepository) GetAttachments(postUid uint) ([]models.BoardAttachment, error) {
	var items []models.BoardAttachment
	query := fmt.Sprintf("SELECT uid, name, path FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_FILE)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		item := models.BoardAttachment{}
		err = rows.Scan(&item.Uid, &item.Name, &path)
		if err != nil {
			return nil, err
		}
		item.Size = utils.GetFileSize(path)
		if item.Size < 1 {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// 게시글에 첨부된 이미지들 가져오기
func (r *TsboardBoardViewRepository) GetAttachedImages(postUid uint) ([]models.BoardAttachedImage, error) {
	var items []models.BoardAttachedImage
	query := fmt.Sprintf("SELECT uid, path FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_FILE)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fileUid uint
		var filePath string
		err = rows.Scan(&fileUid, &filePath)
		if err != nil {
			return nil, err
		}

		thumb := r.GetThumbnailImage(fileUid)
		exif := r.GetExif(fileUid)
		desc := r.GetImageDescription(fileUid)
		item := models.BoardAttachedImage{
			File: models.BoardFile{
				Uid:  fileUid,
				Path: filePath,
			},
			Thumbnail: models.BoardThumbnail{
				Small: thumb.Small,
				Large: thumb.Large,
			},
			Exif:        exif,
			Description: desc,
		}
		items = append(items, item)
	}
	return items, nil
}

// 게시판 기본 설정값들만 가져오기
func (r *TsboardBoardViewRepository) GetBasicBoardConfig(boardUid uint) models.BoardBasicConfig {
	query := fmt.Sprintf("SELECT id, type, name FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	result := models.BoardBasicConfig{}
	r.db.QueryRow(query, boardUid).Scan(&result.Id, &result.Type, &result.Name)
	return result
}

// 첨부파일 다운로드에 필요한 정보 가져오기
func (r *TsboardBoardViewRepository) GetDownloadInfo(fileUid uint) models.BoardViewDownloadResult {
	var result models.BoardViewDownloadResult
	query := fmt.Sprintf("SELECT name, path FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE)
	r.db.QueryRow(query, fileUid).Scan(&result.Name, &result.Path)
	return result
}

// EXIF 정보 가져오기
func (r *TsboardBoardViewRepository) GetExif(fileUid uint) models.BoardExif {
	exif := models.BoardExif{}
	query := fmt.Sprintf(`SELECT make, model, aperture, iso, focal_length, exposure, width, height, date 
												FROM %s%s WHERE file_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_EXIF)
	r.db.QueryRow(query, fileUid).Scan(&exif.Make, &exif.Model, &exif.Aperture, &exif.ISO, &exif.FocalLength, &exif.Exposure, &exif.Width, &exif.Height, &exif.Date)
	return exif
}

// 생성된 이미지 설명글 가져오기
func (r *TsboardBoardViewRepository) GetImageDescription(fileUid uint) string {
	var description string
	query := fmt.Sprintf("SELECT description FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE_DESC)
	r.db.QueryRow(query, fileUid).Scan(&description)
	return description
}

// Action에 필요한 포인트 양 확인하기
func (r *TsboardBoardViewRepository) GetNeededLevelPoint(boardUid uint, action models.BoardAction) (int, int) {
	var level, point int
	act := action.String()
	query := fmt.Sprintf("SELECT level_%s, point_%s FROM %s%s WHERE uid = ? LIMIT 1",
		act, act, configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&level, &point)
	return level, point
}

// 현재 게시글의 이전 게시글 번호 가져오기
func (r *TsboardBoardViewRepository) GetPrevPostUid(boardUid uint, postUid uint) uint {
	var prevUid uint
	query := fmt.Sprintf(`SELECT uid FROM %s%s WHERE board_uid = ? AND status != ? AND uid < ? 
												ORDER BY uid DESC LIMIT 1`, configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, boardUid, models.CONTENT_REMOVED, postUid).Scan(&prevUid)
	return prevUid
}

// 현재 게시글의 다음 게시글 번호 가져오기
func (r *TsboardBoardViewRepository) GetNextPostUid(boardUid uint, postUid uint) uint {
	var nextUid uint
	query := fmt.Sprintf(`SELECT uid FROM %s%s WHERE board_uid = ? AND status != ? AND uid > ?
											 ORDER BY uid ASC LIMIT 1`, configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, boardUid, models.CONTENT_REMOVED, postUid).Scan(&nextUid)
	return nextUid
}

// 게시글 보기 시 게시글 내용 가져오기
func (r *TsboardBoardViewRepository) GetPost(postUid uint, actionUserUid uint) (models.BoardListItem, error) {
	item := models.BoardListItem{}
	var writerUid uint
	query := fmt.Sprintf("SELECT %s FROM %s%s WHERE uid = ? AND status != ? LIMIT 1",
		POST_COLUMNS, configs.Env.Prefix, models.TABLE_POST)
	err := r.db.QueryRow(query, postUid, models.CONTENT_REMOVED).Scan(&item.Uid, &writerUid, &item.Category.Uid, &item.Title, &item.Content, &item.Submitted, &item.Modified, &item.Hit, &item.Status)
	if err != nil {
		return item, err
	}

	item.Writer = r.board.GetWriterInfo(writerUid)
	item.Like = r.board.GetCountByTable(models.TABLE_POST_LIKE, postUid)
	item.Liked = r.board.CheckLikedPost(postUid, actionUserUid)
	item.Category = r.board.GetCategoryByUid(item.Category.Uid)
	item.Comment = r.board.GetCountByTable(models.TABLE_COMMENT, postUid)
	item.Cover = r.board.GetCoverImage(postUid)
	return item, nil
}

// 게시글에 등록된 해시태그들 가져오기
func (r *TsboardBoardViewRepository) GetTags(postUid uint) []models.Pair {
	var items []models.Pair
	query := fmt.Sprintf("SELECT hashtag_uid FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return items
	}
	rows.Close()

	for rows.Next() {
		item := models.Pair{}
		err = rows.Scan(&item.Uid)
		if err != nil {
			return items
		}
		item.Name = r.GetTagName(item.Uid)
		items = append(items, item)
	}
	return items
}

// 해시태그명 가져오기
func (r *TsboardBoardViewRepository) GetTagName(hashtagUid uint) string {
	var name string
	query := fmt.Sprintf("SELECT name FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_HASHTAG)
	r.db.QueryRow(query, hashtagUid).Scan(&name)
	return name
}

// 썸네일 이미지 가져오기
func (r *TsboardBoardViewRepository) GetThumbnailImage(fileUid uint) models.BoardThumbnail {
	thumb := models.BoardThumbnail{}
	query := fmt.Sprintf("SELECT path, full_path FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE_THUMB)
	r.db.QueryRow(query, fileUid).Scan(&thumb.Small, &thumb.Large)
	return thumb
}

// 게시글 작성자의 최근 댓글들 가져오기
func (r *TsboardBoardViewRepository) GetWriterLatestComment(writerUid uint, limit uint) ([]models.BoardWriterLatestComment, error) {
	query := fmt.Sprintf(`SELECT uid, board_uid, post_uid, content, submitted 
												FROM %s%s WHERE user_uid = ? AND status != ? 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_COMMENT)
	rows, err := r.db.Query(query, writerUid, models.CONTENT_REMOVED, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.BoardWriterLatestComment
	for rows.Next() {
		item := models.BoardWriterLatestComment{}
		var uid, boardUid uint
		err = rows.Scan(&uid, &boardUid, &item.PostUid, &item.Content, &item.Submitted)
		if err != nil {
			return nil, err
		}
		item.Board = r.GetBasicBoardConfig(boardUid)
		item.Like = r.board.GetCountByTable(models.TABLE_COMMENT_LIKE, item.PostUid)
		items = append(items, item)
	}
	return items, nil
}

// 게시글 작성자의 최근 포스트들 가져오기
func (r *TsboardBoardViewRepository) GetWriterLatestPost(writerUid uint, limit uint) ([]models.BoardWriterLatestPost, error) {
	query := fmt.Sprintf(`SELECT uid, board_uid, title, submitted FROM %s%s 
												WHERE user_uid = ? AND status != ? 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, writerUid, models.CONTENT_REMOVED, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.BoardWriterLatestPost
	for rows.Next() {
		item := models.BoardWriterLatestPost{}
		var boardUid uint
		err = rows.Scan(&item.PostUid, &boardUid, &item.Title, &item.Submitted)
		if err != nil {
			return nil, err
		}
		item.Board = r.GetBasicBoardConfig(boardUid)
		item.Comment = r.board.GetCountByTable(models.TABLE_COMMENT, item.PostUid)
		item.Like = r.board.GetCountByTable(models.TABLE_COMMENT_LIKE, item.PostUid)
		items = append(items, item)
	}
	return items, nil
}

// 게시글에 대해 좋아요를 클릭한 적 있는지 확인
func (r *TsboardBoardViewRepository) IsLikedPost(postUid uint, actionUserUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT post_uid FROM %s%s WHERE post_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST_LIKE)
	r.db.QueryRow(query, postUid, actionUserUid).Scan(&uid)
	return uid > 0
}

// 게시글 혹은 댓글 작성자인지 확인
func (r *TsboardBoardViewRepository) IsWriter(table models.Table, targetUid uint, userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, table)
	r.db.QueryRow(query, targetUid).Scan(&uid)
	return uid == userUid
}

// 게시글에 대한 좋아요를 추가하기
func (r *TsboardBoardViewRepository) InsertLikePost(param models.BoardViewLikeParameter) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, post_uid, user_uid, liked, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_POST_LIKE)
	r.db.Exec(query, param.BoardUid, param.PostUid, param.UserUid, param.Liked, time.Now().UnixMilli())
}

// 첨부파일 및 썸네일들 삭제하기
func (r *TsboardBoardViewRepository) RemoveAttachments(postUid uint) []string {
	var removes []string
	query := fmt.Sprintf("SELECT uid, path FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_FILE)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return removes
	}
	defer rows.Close()

	for rows.Next() {
		var fileUid uint
		var filePath string
		rows.Scan(&fileUid, &filePath)
		attachs := r.RemoveAttachedFile(fileUid, filePath)
		removes = append(removes, attachs...)
	}
	return removes
}

// 첨부파일 삭제
func (r *TsboardBoardViewRepository) RemoveAttachedFile(fileUid uint, filePath string) []string {
	var removes []string
	removes = append(removes, filePath)

	isImg := utils.IsImage(filePath)
	if isImg {
		thumbs := r.RemoveThumbnails(fileUid)
		removes = append(removes, thumbs...)

		query := fmt.Sprintf("DELETE FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE_DESC)
		r.db.Exec(query, fileUid)
		query = fmt.Sprintf("DELETE FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_EXIF)
		r.db.Exec(query, fileUid)
	}
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE)
	r.db.Exec(query, fileUid)
	return removes
}

// 게시글 삭제 상태로 변경하기
func (r *TsboardBoardViewRepository) RemovePost(postUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	r.db.Exec(query, models.CONTENT_REMOVED, postUid)
	query = fmt.Sprintf("UPDATE %s%s SET status = ? WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.Exec(query, models.CONTENT_REMOVED, postUid)
}

// 게시글에 등록된 태그 제거하기
func (r *TsboardBoardViewRepository) RemovePostTags(postUid uint) {
	query := fmt.Sprintf("SELECT hashtag_uid FROM %s%s WHERE post_uid = ?",
		configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		hashtagUid := 0
		rows.Scan(&hashtagUid)
		query = fmt.Sprintf("UPDATE %s%s SET used = used - 1 WHERE uid = ? LIMIT 1",
			configs.Env.Prefix, models.TABLE_POST_HASHTAG)
		r.db.Exec(query, hashtagUid)
	}
	query = fmt.Sprintf("DELETE FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	r.db.Exec(query, postUid)
}

// 썸네일 삭제하기
func (r *TsboardBoardViewRepository) RemoveThumbnails(fileUid uint) []string {
	var uid uint
	var path, fullPath string
	var removes []string
	query := fmt.Sprintf("SELECT uid, path, full_path FROM %s%s WHERE file_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)
	r.db.QueryRow(query, fileUid).Scan(&uid, &path, &fullPath)
	removes = []string{path, fullPath}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE_THUMB)
	r.db.Exec(query, uid)
	return removes
}

// 게시글에 대한 좋아요를 변경하기
func (r *TsboardBoardViewRepository) UpdateLikePost(param models.BoardViewLikeParameter) {
	query := fmt.Sprintf(`UPDATE %s%s SET liked = ?, timestamp = ? 
												WHERE post_uid = ? AND user_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_POST_LIKE)
	r.db.Exec(query, param.Liked, time.Now().UnixMilli(), param.PostUid, param.UserUid)
}

// 조회수 업데이트 하기
func (r *TsboardBoardViewRepository) UpdatePostHit(postUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET hit = hit + 1 WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	r.db.Exec(query, postUid)
}

// 게시글의 소속 게시판 변경하기
func (r *TsboardBoardViewRepository) UpdatePostBoardUid(targetBoardUid uint, postUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET board_uid = ?, modified = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)
	r.db.Exec(query, targetBoardUid, time.Now().UnixMilli(), postUid)
}

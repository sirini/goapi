package repositories

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
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
	GetPostItem(postUid uint, actionUserUid uint) (models.BoardListItem, error)
	GetTags(postUid uint) []models.Pair
	GetTagName(hashtagUid uint) string
	GetThumbnailImage(fileUid uint) models.BoardThumbnail
	GetWriterLatestComment(writerUid uint, limit uint) ([]models.BoardWriterLatestComment, error)
	GetWriterLatestPost(writerUid uint, limit uint) ([]models.BoardWriterLatestPost, error)
	InsertLikePost(param models.BoardViewLikeParam)
	IsLikedPost(postUid uint, actionUserUid uint) bool
	IsWriter(table models.Table, targetUid uint, userUid uint) bool
	RemoveAttachments(postUid uint) []string
	RemoveAttachedFile(fileUid uint, filePath string) []string
	RemoveComments(postUid uint)
	RemoveExif(fileUid uint)
	RemoveImageDescription(fileUid uint)
	RemovePost(postUid uint) error
	RemovePostTags(postUid uint)
	RemoveThumbnails(fileUid uint) []string
	UpdateLikePost(param models.BoardViewLikeParam)
	UpdatePostHit(postUid uint)
	UpdatePostBoardUid(targetBoardUid uint, postUid uint)
}

type NuboBoardViewRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewNuboBoardViewRepository(db *sql.DB, board BoardRepository) *NuboBoardViewRepository {
	return &NuboBoardViewRepository{db: db, board: board}
}

// 글작성자에게 차단당한 사용자인지 확인하기
func (r *NuboBoardViewRepository) CheckBannedByWriter(postUid uint, viewerUid uint) bool {
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
func (r *NuboBoardViewRepository) GetAllBoards() []models.BoardItem {
	items := make([]models.BoardItem, 0)
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
func (r *NuboBoardViewRepository) GetAttachments(postUid uint) ([]models.BoardAttachment, error) {
	items := make([]models.BoardAttachment, 0)
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
		items = append(items, item)
	}
	return items, nil
}

// 게시글에 첨부된 이미지들 가져오기
func (r *NuboBoardViewRepository) GetAttachedImages(postUid uint) ([]models.BoardAttachedImage, error) {
	items := make([]models.BoardAttachedImage, 0)
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

		imageExtensions := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
			".bmp":  true,
			".tiff": true,
			".webp": true,
		}
		ext := filepath.Ext(filePath)
		lowerExt := strings.ToLower(ext)
		if !imageExtensions[lowerExt] {
			continue
		}

		imgSize := utils.GetFileSize(filePath)
		if imgSize < 1 {
			continue
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
func (r *NuboBoardViewRepository) GetBasicBoardConfig(boardUid uint) models.BoardBasicConfig {
	result := models.BoardBasicConfig{}
	query := fmt.Sprintf("SELECT id, type, name FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&result.Id, &result.Type, &result.Name)
	return result
}

// 첨부파일 다운로드에 필요한 정보 가져오기
func (r *NuboBoardViewRepository) GetDownloadInfo(fileUid uint) models.BoardViewDownloadResult {
	var result models.BoardViewDownloadResult
	query := fmt.Sprintf("SELECT name, path FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE)

	r.db.QueryRow(query, fileUid).Scan(&result.Name, &result.Path)
	return result
}

// EXIF 정보 가져오기
func (r *NuboBoardViewRepository) GetExif(fileUid uint) models.BoardExif {
	exif := models.BoardExif{}
	query := fmt.Sprintf(`SELECT make, model, aperture, iso, focal_length, exposure, width, height, date 
												FROM %s%s WHERE file_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_EXIF)

	r.db.QueryRow(query, fileUid).Scan(&exif.Make, &exif.Model, &exif.Aperture, &exif.ISO, &exif.FocalLength, &exif.Exposure, &exif.Width, &exif.Height, &exif.Date)
	return exif
}

// 생성된 이미지 설명글 가져오기
func (r *NuboBoardViewRepository) GetImageDescription(fileUid uint) string {
	var description string
	query := fmt.Sprintf("SELECT description FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE_DESC)

	r.db.QueryRow(query, fileUid).Scan(&description)
	return description
}

// Action에 필요한 포인트 양 확인하기
func (r *NuboBoardViewRepository) GetNeededLevelPoint(boardUid uint, action models.BoardAction) (int, int) {
	var level, point int
	act := action.String()
	query := fmt.Sprintf("SELECT level_%s, point_%s FROM %s%s WHERE uid = ? LIMIT 1",
		act, act, configs.Env.Prefix, models.TABLE_BOARD)

	r.db.QueryRow(query, boardUid).Scan(&level, &point)
	return level, point
}

// 현재 게시글의 이전 게시글 번호 가져오기
func (r *NuboBoardViewRepository) GetPrevPostUid(boardUid uint, postUid uint) uint {
	var prevUid uint
	query := fmt.Sprintf(`SELECT uid FROM %s%s WHERE board_uid = ? AND status != ? AND uid < ? 
												ORDER BY uid DESC LIMIT 1`, configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, boardUid, models.CONTENT_REMOVED, postUid).Scan(&prevUid)
	return prevUid
}

// 현재 게시글의 다음 게시글 번호 가져오기
func (r *NuboBoardViewRepository) GetNextPostUid(boardUid uint, postUid uint) uint {
	var nextUid uint
	query := fmt.Sprintf(`SELECT uid FROM %s%s WHERE board_uid = ? AND status != ? AND uid > ?
											 ORDER BY uid ASC LIMIT 1`, configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, boardUid, models.CONTENT_REMOVED, postUid).Scan(&nextUid)
	return nextUid
}

// 게시글 보기 시 글 내용 가져오기
func (r *NuboBoardViewRepository) GetPostItem(postUid uint, actionUserUid uint) (models.BoardListItem, error) {
	item := models.BoardListItem{}
	prefix := configs.Env.Prefix

	query := fmt.Sprintf(`SELECT p.uid, p.user_uid, p.category_uid, p.title, p.content, p.submitted, p.modified, p.hit, p.status,
			u.name, u.profile,
			c.name,
			COALESCE((SELECT path FROM %s%s WHERE post_uid = p.uid LIMIT 1), ''),
			(SELECT COUNT(*) FROM %s%s WHERE post_uid = p.uid AND status != ?),
			(SELECT COUNT(*) FROM %s%s WHERE post_uid = p.uid AND liked = 1),
			EXISTS(SELECT 1 FROM %s%s WHERE post_uid = p.uid AND user_uid = ? AND liked = 1)
		FROM %s%s AS p
		LEFT JOIN %s%s AS u ON p.user_uid = u.uid
		LEFT JOIN %s%s AS c ON p.category_uid = c.uid
		WHERE p.uid = ? AND p.status != ?`,
		prefix, models.TABLE_FILE_THUMB,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_POST_LIKE,
		prefix, models.TABLE_POST_LIKE,
		prefix, models.TABLE_POST,
		prefix, models.TABLE_USER,
		prefix, models.TABLE_BOARD_CAT,
	)

	err := r.db.QueryRow(query,
		models.CONTENT_REMOVED,
		actionUserUid,
		postUid,
		models.CONTENT_REMOVED,
	).Scan(
		&item.Uid,
		&item.Writer.UserUid,
		&item.Category.Uid,
		&item.Title,
		&item.Content,
		&item.Submitted,
		&item.Modified,
		&item.Hit,
		&item.Status,
		&item.Writer.Name,
		&item.Writer.Profile,
		&item.Category.Name,
		&item.Cover,
		&item.Comment,
		&item.Like,
		&item.Liked,
	)
	if err != nil {
		return item, err
	}
	return item, nil
}

// 게시글에 등록된 해시태그들 가져오기
func (r *NuboBoardViewRepository) GetTags(postUid uint) []models.Pair {
	items := make([]models.Pair, 0)
	query := fmt.Sprintf("SELECT hashtag_uid FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return items
	}
	defer rows.Close()

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
func (r *NuboBoardViewRepository) GetTagName(hashtagUid uint) string {
	var name string
	query := fmt.Sprintf("SELECT name FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_HASHTAG)
	r.db.QueryRow(query, hashtagUid).Scan(&name)
	return name
}

// 썸네일 이미지 가져오기
func (r *NuboBoardViewRepository) GetThumbnailImage(fileUid uint) models.BoardThumbnail {
	thumb := models.BoardThumbnail{}
	query := fmt.Sprintf("SELECT path, full_path FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE_THUMB)

	r.db.QueryRow(query, fileUid).Scan(&thumb.Small, &thumb.Large)
	return thumb
}

// 게시글 작성자의 최근 댓글들 가져오기
func (r *NuboBoardViewRepository) GetWriterLatestComment(writerUid uint, limit uint) ([]models.BoardWriterLatestComment, error) {
	query := fmt.Sprintf(`SELECT uid, board_uid, post_uid, content, submitted 
												FROM %s%s WHERE user_uid = ? AND status != ? 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_COMMENT)

	rows, err := r.db.Query(query, writerUid, models.CONTENT_REMOVED, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.BoardWriterLatestComment, 0)
	for rows.Next() {
		item := models.BoardWriterLatestComment{}
		var uid, boardUid uint
		err = rows.Scan(&uid, &boardUid, &item.PostUid, &item.Content, &item.Submitted)
		if err != nil {
			return nil, err
		}
		item.Board = r.GetBasicBoardConfig(boardUid)
		item.Like = r.board.GetCommentLikeCount(item.PostUid)
		items = append(items, item)
	}
	return items, nil
}

// 게시글 작성자의 최근 포스트들 가져오기
func (r *NuboBoardViewRepository) GetWriterLatestPost(writerUid uint, limit uint) ([]models.BoardWriterLatestPost, error) {
	query := fmt.Sprintf(`SELECT uid, board_uid, title, submitted FROM %s%s 
												WHERE user_uid = ? AND status != ? 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_POST)

	rows, err := r.db.Query(query, writerUid, models.CONTENT_REMOVED, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.BoardWriterLatestPost, 0)
	for rows.Next() {
		item := models.BoardWriterLatestPost{}
		var boardUid uint
		err = rows.Scan(&item.PostUid, &boardUid, &item.Title, &item.Submitted)
		if err != nil {
			return nil, err
		}
		item.Board = r.GetBasicBoardConfig(boardUid)
		item.Comment = r.board.GetCommentCount(item.PostUid)
		item.Like = r.board.GetLikeCount(item.PostUid)
		items = append(items, item)
	}
	return items, nil
}

// 게시글에 대해 좋아요를 클릭한 적 있는지 확인
func (r *NuboBoardViewRepository) IsLikedPost(postUid uint, actionUserUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT post_uid FROM %s%s WHERE post_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST_LIKE)

	r.db.QueryRow(query, postUid, actionUserUid).Scan(&uid)
	return uid > 0
}

// 게시글 혹은 댓글 작성자인지 확인
func (r *NuboBoardViewRepository) IsWriter(table models.Table, targetUid uint, userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, table)
	r.db.QueryRow(query, targetUid).Scan(&uid)
	return uid == userUid
}

// 게시글에 대한 좋아요를 추가하기
func (r *NuboBoardViewRepository) InsertLikePost(param models.BoardViewLikeParam) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, post_uid, user_uid, liked, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_POST_LIKE)

	r.db.Exec(query, param.BoardUid, param.PostUid, param.UserUid, param.Liked, time.Now().UnixMilli())
}

// 첨부파일 및 썸네일들 삭제하기
func (r *NuboBoardViewRepository) RemoveAttachments(postUid uint) []string {
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
func (r *NuboBoardViewRepository) RemoveAttachedFile(fileUid uint, filePath string) []string {
	var removes []string
	removes = append(removes, filePath)

	isImg := utils.IsImage(filePath)
	if isImg {
		thumbs := r.RemoveThumbnails(fileUid)
		removes = append(removes, thumbs...)

		r.RemoveImageDescription(fileUid)
		r.RemoveExif(fileUid)
	}
	query := fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE)
	r.db.Exec(query, fileUid)
	return removes
}

// 게시글에 등록된 댓글들 삭제 처리하기
func (r *NuboBoardViewRepository) RemoveComments(postUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.Exec(query, models.CONTENT_REMOVED, postUid)
}

// EXIF 삭제
func (r *NuboBoardViewRepository) RemoveExif(fileUid uint) {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_EXIF)
	r.db.Exec(query, fileUid)
}

// AI로 생성한 이미지 설명글 삭제
func (r *NuboBoardViewRepository) RemoveImageDescription(fileUid uint) {
	query := fmt.Sprintf("DELETE FROM %s%s WHERE file_uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE_DESC)
	r.db.Exec(query, fileUid)
}

// 게시글 삭제 상태로 변경하기
func (r *NuboBoardViewRepository) RemovePost(postUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	_, err := r.db.Exec(query, models.CONTENT_REMOVED, postUid)
	return err
}

// 게시글에 등록된 태그 제거하기
func (r *NuboBoardViewRepository) RemovePostTags(postUid uint) {
	query := fmt.Sprintf("SELECT hashtag_uid FROM %s%s WHERE post_uid = ?",
		configs.Env.Prefix, models.TABLE_POST_HASHTAG)

	rows, err := r.db.Query(query, postUid)
	if err != nil {
		return
	}
	defer rows.Close()

	query = fmt.Sprintf("UPDATE %s%s SET used = used - 1 WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_HASHTAG)
	stmtUpdate, err := r.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmtUpdate.Close()

	for rows.Next() {
		hashtagUid := 0
		rows.Scan(&hashtagUid)
		stmtUpdate.Exec(hashtagUid)
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	r.db.Exec(query, postUid)
}

// 썸네일 삭제하기
func (r *NuboBoardViewRepository) RemoveThumbnails(fileUid uint) []string {
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
func (r *NuboBoardViewRepository) UpdateLikePost(param models.BoardViewLikeParam) {
	query := fmt.Sprintf(`UPDATE %s%s SET liked = ?, timestamp = ? 
												WHERE post_uid = ? AND user_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_POST_LIKE)

	r.db.Exec(query, param.Liked, time.Now().UnixMilli(), param.PostUid, param.UserUid)
}

// 조회수 업데이트 하기
func (r *NuboBoardViewRepository) UpdatePostHit(postUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET hit = hit + 1 WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_POST)
	r.db.Exec(query, postUid)
}

// 게시글의 소속 게시판 변경하기
func (r *NuboBoardViewRepository) UpdatePostBoardUid(targetBoardUid uint, postUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET board_uid = ?, modified = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)

	r.db.Exec(query, targetBoardUid, time.Now().UnixMilli(), postUid)
}

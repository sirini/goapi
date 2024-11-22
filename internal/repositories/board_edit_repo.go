package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardEditRepository interface {
	CheckWriterForBlog(boardUid uint, actionUserUid uint) bool
	FindTagUidByName(name string) uint
	GetInsertedImages(param models.EditorInsertImageParameter) ([]models.Pair, error)
	GetMaxImageUid(boardUid uint, actionUserUid uint) uint
	GetSuggestionTags(input string, bunch uint) []models.EditorTagItem
	GetTotalImageCount(boardUid uint, actionUserUid uint) uint
	InsertExif(fileUid uint, postUid uint, exif models.BoardExif)
	InsertFile(param models.EditorSaveFileParameter) uint
	InsertFileThumbnail(param models.EditorSaveThumbnailParameter)
	InsertImageDescription(fileUid uint, postUid uint, description string)
	InsertImagePaths(boardUid uint, userUid uint, paths []string)
	InsertPost(param models.EditorWriteParameter) uint
	InsertPostHashtag(boardUid uint, postUid uint, hashtagUid uint)
	InsertTag(boardUid uint, postUid uint, tag string) uint
	RemoveInsertedImage(imageUid uint, actionUserUid uint) string
	UpdateTag(hashtagUid uint)
}

type TsboardBoardEditRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardBoardEditRepository(db *sql.DB, board BoardRepository) *TsboardBoardEditRepository {
	return &TsboardBoardEditRepository{db: db, board: board}
}

// 블로그에 글을 남기는 경우에는 작성자가 블로그 주인(=게시판 관리자)인지 확인
func (r *TsboardBoardEditRepository) CheckWriterForBlog(boardUid uint, actionUserUid uint) bool {
	var adminUid uint
	var boardType uint8
	query := fmt.Sprintf("SELECT admin_uid, type FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&adminUid, &boardType)
	return boardType != uint8(models.BOARD_BLOG) || actionUserUid == adminUid
}

// 태그명에 해당하는 고유 번호 반환하기
func (r *TsboardBoardEditRepository) FindTagUidByName(name string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1", configs.Env.Prefix, models.TABLE_HASHTAG)
	r.db.QueryRow(query, name).Scan(&uid)
	return uid
}

// 게시글에 삽입했던 이미지들 가져오기
func (r *TsboardBoardEditRepository) GetInsertedImages(param models.EditorInsertImageParameter) ([]models.Pair, error) {
	var images []models.Pair
	if param.LastUid < 1 {
		param.LastUid = r.GetMaxImageUid(param.BoardUid, param.UserUid) + 1
	}
	query := fmt.Sprintf(`SELECT uid, path FROM %s%s WHERE uid < ? AND board_uid = ? AND user_uid = ? 
												ORDER BY uid DESC LIMIT ?`, configs.Env.Prefix, models.TABLE_IMAGE)
	rows, err := r.db.Query(query, param.LastUid, param.BoardUid, param.UserUid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		image := models.Pair{}
		rows.Scan(&image.Uid, &image.Name)
		images = append(images, image)
	}
	return images, nil
}

// 내가 올린 이미지들의 가장 최근 고유 번호 반환
func (r *TsboardBoardEditRepository) GetMaxImageUid(boardUid uint, actionUserUid uint) uint {
	var uid uint
	query := fmt.Sprintf("SELECT MAX(uid) FROM %s%s WHERE board_uid = ? AND user_uid = ?",
		configs.Env.Prefix, models.TABLE_IMAGE)
	r.db.QueryRow(query, boardUid, actionUserUid).Scan(&uid)
	return uid
}

// 태그 추천하기 목록 가져오기
func (r *TsboardBoardEditRepository) GetSuggestionTags(input string, bunch uint) []models.EditorTagItem {
	items := make([]models.EditorTagItem, 0)
	query := fmt.Sprintf("SELECT uid, name, used FROM %s%s WHERE name LIKE ? LIMIT ?",
		configs.Env.Prefix, models.TABLE_HASHTAG)
	rows, err := r.db.Query(query, "%"+input+"%", bunch)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.EditorTagItem{}
		rows.Scan(&item.Uid, &item.Name, &item.Count)
		items = append(items, item)
	}
	return items
}

// 내가 올린 이미지 총 갯수 반환
func (r *TsboardBoardEditRepository) GetTotalImageCount(boardUid uint, actionUserUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE board_uid = ? AND user_uid = ?",
		configs.Env.Prefix, models.TABLE_IMAGE)
	r.db.QueryRow(query, boardUid, actionUserUid).Scan(&count)
	return count
}

// EXIF 정보 저장하기
func (r *TsboardBoardEditRepository) InsertExif(fileUid uint, postUid uint, exif models.BoardExif) {
	query := fmt.Sprintf(`INSERT INTO %s%s (
		file_uid, post_uid, make, model, aperture, iso, focal_length, exposure, width, height, date) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_EXIF)
	r.db.Exec(query, fileUid, postUid,
		exif.Make, exif.Model, exif.Aperture, exif.ISO, exif.FocalLength,
		exif.Exposure, exif.Width, exif.Height, exif.Date)
}

// 첨부파일 경로 저장하기
func (r *TsboardBoardEditRepository) InsertFile(param models.EditorSaveFileParameter) uint {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, post_uid, name, path, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_FILE)
	result, _ := r.db.Exec(query, param.BoardUid, param.PostUid, param.Name, param.Path, time.Now().UnixMilli())
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 썸네일 경로 저장하기
func (r *TsboardBoardEditRepository) InsertFileThumbnail(param models.EditorSaveThumbnailParameter) {
	query := fmt.Sprintf("INSERT INTO %s%s (file_uid, post_uid, path, full_path) VALUES (?, ?, ?, ?)",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)
	r.db.Exec(query, param.FileUid, param.PostUid, param.Small, param.Large)
}

// 이미지 설명글 저장하기 (OpenAI API 사용 시에만 가능)
func (r *TsboardBoardEditRepository) InsertImageDescription(fileUid uint, postUid uint, description string) {
	query := fmt.Sprintf("INSERT INTO %s%s (file_uid, post_uid, description) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_IMAGE_DESC)
	r.db.Exec(query, fileUid, postUid, description)
}

// 게시글에 삽입한 이미지 정보들을 한 번에 저장하기
func (r *TsboardBoardEditRepository) InsertImagePaths(boardUid uint, userUid uint, paths []string) {
	query := fmt.Sprintf("INSERT INTO %s%s (board_uid, user_uid, path, timestamp) VALUES ",
		configs.Env.Prefix, models.TABLE_IMAGE)
	values := make([]interface{}, 0)
	now := time.Now().UnixMilli()

	for _, path := range paths {
		query += "(?, ?, ?, ?),"
		values = append(values, boardUid, userUid, path, now)
	}

	query = query[:len(query)-1]
	r.db.Exec(query, values...)
}

// 새 게시글 작성하기
func (r *TsboardBoardEditRepository) InsertPost(param models.EditorWriteParameter) uint {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(board_uid, user_uid, category_uid, title, content, submitted, modified, hit, status) 
												VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_POST)
	status := models.CONTENT_NORMAL
	if param.IsNotice {
		status = models.CONTENT_NOTICE
	} else if param.IsSecret {
		status = models.CONTENT_SECRET
	}

	result, _ := r.db.Exec(
		query,
		param.BoardUid,
		param.UserUid,
		param.CategoryUid,
		param.Title,
		param.Content,
		time.Now().UnixMilli(),
		0,
		0,
		status,
	)

	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(insertId)
}

// 해시태그와 게시글 번호 연결 정보 저장하기
func (r *TsboardBoardEditRepository) InsertPostHashtag(boardUid uint, postUid uint, hashtagUid uint) {
	query := fmt.Sprintf("INSERT INTO %s%s (board_uid, post_uid, hashtag_uid) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	r.db.Exec(query, boardUid, postUid, hashtagUid)
}

// 신규 태그 저장하기
func (r *TsboardBoardEditRepository) InsertTag(boardUid uint, postUid uint, tag string) uint {
	query := fmt.Sprintf("INSERT INTO %s%s (name, used, timestamp) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_HASHTAG)
	result, _ := r.db.Exec(query, tag, 1, time.Now().UnixMilli())
	hashtagUid, err := result.LastInsertId()
	if err != nil {
		return models.FAILED
	}
	return uint(hashtagUid)
}

// 게시글에 삽입한 이미지 삭제하기
func (r *TsboardBoardEditRepository) RemoveInsertedImage(imageUid uint, actionUserUid uint) string {
	query := fmt.Sprintf("SELECT user_uid, path FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_IMAGE)
	var userUid uint
	var path string
	r.db.QueryRow(query, imageUid).Scan(&userUid, &path)

	if actionUserUid != userUid {
		return ""
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE)
	r.db.Exec(query, imageUid)
	return path
}

// 기존 태그 사용 횟수 올리고 태그와 게시글 번호 연결하기
func (r *TsboardBoardEditRepository) UpdateTag(hashtagUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET used = used + 1 WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_HASHTAG)
	r.db.Exec(query, hashtagUid)
}

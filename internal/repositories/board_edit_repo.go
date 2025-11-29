package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardEditRepository interface {
	CheckWriterForBlog(boardUid uint, actionUserUid uint) (bool, error)
	FindTagUidByName(name string) uint
	FindAttachedPathByUid(fileUid uint) (string, error)
	GetInsertedImages(param models.EditorInsertImageParameter) ([]models.Pair, error)
	GetMaxImageUid(boardUid uint, actionUserUid uint) (uint, error)
	GetSuggestionTitles(input string, bunch uint) ([]string, error)
	GetSuggestionTags(input string, bunch uint) ([]models.EditorTagItem, error)
	GetTotalImageCount(boardUid uint, actionUserUid uint) (uint, error)
	InsertExif(fileUid uint, postUid uint, exif models.BoardExif) error
	InsertFile(param models.EditorSaveFileParameter) (uint, error)
	InsertFileThumbnail(param models.EditorSaveThumbnailParameter) error
	InsertImageDescription(fileUid uint, postUid uint, description string) error
	InsertImagePaths(boardUid uint, userUid uint, paths []string) error
	InsertPost(param models.EditorWriteParameter) (uint, error)
	InsertPostHashtag(boardUid uint, postUid uint, hashtagUid uint) error
	InsertTag(boardUid uint, postUid uint, tag string) (uint, error)
	RemoveInsertedImage(imageUid uint, actionUserUid uint) (string, error)
	UpdatePost(param models.EditorModifyParameter) error
	UpdateTag(hashtagUid uint) error
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
func (r *TsboardBoardEditRepository) CheckWriterForBlog(boardUid uint, actionUserUid uint) (bool, error) {
	var adminUid uint
	var boardType uint8
	query := fmt.Sprintf("SELECT admin_uid, type FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD)
	err := r.db.QueryRow(query, boardUid).Scan(&adminUid, &boardType)
	return boardType != uint8(models.BOARD_BLOG) || actionUserUid == adminUid, err
}

// 게시글 수정에서 삭제할 파일의 경로 가져오기
func (r *TsboardBoardEditRepository) FindAttachedPathByUid(fileUid uint) (string, error) {
	var path string
	query := fmt.Sprintf("SELECT path FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_FILE)
	err := r.db.QueryRow(query, fileUid).Scan(&path)
	return path, err
}

// 태그명에 해당하는 고유 번호 반환하기
func (r *TsboardBoardEditRepository) FindTagUidByName(name string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1", configs.Env.Prefix, models.TABLE_HASHTAG)
	r.db.QueryRow(query, name).Scan(&uid) // 기존 태그 없으면 무시
	return uid
}

// 게시글에 삽입했던 이미지들 가져오기
func (r *TsboardBoardEditRepository) GetInsertedImages(param models.EditorInsertImageParameter) ([]models.Pair, error) {
	images := make([]models.Pair, 0)
	if param.LastUid < 1 {
		maxUid, err := r.GetMaxImageUid(param.BoardUid, param.UserUid)
		if err != nil {
			return nil, err
		}
		param.LastUid = maxUid + 1
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
func (r *TsboardBoardEditRepository) GetMaxImageUid(boardUid uint, actionUserUid uint) (uint, error) {
	var uid uint
	query := fmt.Sprintf("SELECT MAX(uid) FROM %s%s WHERE board_uid = ? AND user_uid = ?",
		configs.Env.Prefix, models.TABLE_IMAGE)
	err := r.db.QueryRow(query, boardUid, actionUserUid).Scan(&uid)
	return uid, err
}

// 유사 제목들 가져오기
func (r *TsboardBoardEditRepository) GetSuggestionTitles(input string, bunch uint) ([]string, error) {
	items := make([]string, 0)
	query := fmt.Sprintf("SELECT title FROM %s%s WHERE title LIKE ? LIMIT ?", configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, "%"+input+"%", bunch)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item string
		rows.Scan(&item)
		items = append(items, item)
	}
	return items, nil
}

// 태그 추천하기 목록 가져오기
func (r *TsboardBoardEditRepository) GetSuggestionTags(input string, bunch uint) ([]models.EditorTagItem, error) {
	items := make([]models.EditorTagItem, 0)
	query := fmt.Sprintf("SELECT uid, name, used FROM %s%s WHERE name LIKE ? LIMIT ?",
		configs.Env.Prefix, models.TABLE_HASHTAG)

	rows, err := r.db.Query(query, "%"+input+"%", bunch)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.EditorTagItem{}
		rows.Scan(&item.Uid, &item.Name, &item.Count)
		items = append(items, item)
	}
	return items, nil
}

// 내가 올린 이미지 총 갯수 반환
func (r *TsboardBoardEditRepository) GetTotalImageCount(boardUid uint, actionUserUid uint) (uint, error) {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE board_uid = ? AND user_uid = ?",
		configs.Env.Prefix, models.TABLE_IMAGE)
	err := r.db.QueryRow(query, boardUid, actionUserUid).Scan(&count)
	return count, err
}

// EXIF 정보 저장하기
func (r *TsboardBoardEditRepository) InsertExif(fileUid uint, postUid uint, exif models.BoardExif) error {
	query := fmt.Sprintf(`INSERT INTO %s%s (
		file_uid, post_uid, make, model, aperture, iso, focal_length, exposure, width, height, date) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_EXIF)

	_, err := r.db.Exec(query, fileUid, postUid,
		exif.Make, exif.Model, exif.Aperture, exif.ISO, exif.FocalLength,
		exif.Exposure, exif.Width, exif.Height, exif.Date)
	return err
}

// 첨부파일 경로 저장하기
func (r *TsboardBoardEditRepository) InsertFile(param models.EditorSaveFileParameter) (uint, error) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, post_uid, name, path, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_FILE)
	result, err := r.db.Exec(query, param.BoardUid, param.PostUid, param.Name, param.Path, time.Now().UnixMilli())
	if err != nil {
		return models.FAILED, err
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED, err
	}
	return uint(insertId), nil
}

// 썸네일 경로 저장하기
func (r *TsboardBoardEditRepository) InsertFileThumbnail(param models.EditorSaveThumbnailParameter) error {
	query := fmt.Sprintf("INSERT INTO %s%s (file_uid, post_uid, path, full_path) VALUES (?, ?, ?, ?)",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)
	_, err := r.db.Exec(query, param.FileUid, param.PostUid, param.Small, param.Large)
	return err
}

// 이미지 설명글 저장하기 (OpenAI API 사용 시에만 가능)
func (r *TsboardBoardEditRepository) InsertImageDescription(fileUid uint, postUid uint, description string) error {
	query := fmt.Sprintf("INSERT INTO %s%s (file_uid, post_uid, description) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_IMAGE_DESC)
	_, err := r.db.Exec(query, fileUid, postUid, description)
	return err
}

// 게시글에 삽입한 이미지 정보들을 한 번에 저장하기
func (r *TsboardBoardEditRepository) InsertImagePaths(boardUid uint, userUid uint, paths []string) error {
	query := fmt.Sprintf("INSERT INTO %s%s (board_uid, user_uid, path, timestamp) VALUES ",
		configs.Env.Prefix, models.TABLE_IMAGE)

	values := make([]interface{}, 0)
	now := time.Now().UnixMilli()

	for _, path := range paths {
		query += "(?, ?, ?, ?),"
		values = append(values, boardUid, userUid, path, now)
	}

	query = query[:len(query)-1]
	_, err := r.db.Exec(query, values...)
	return err
}

// 새 게시글 작성하기
func (r *TsboardBoardEditRepository) InsertPost(param models.EditorWriteParameter) (uint, error) {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(board_uid, user_uid, category_uid, title, content, submitted, modified, hit, status) 
												VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_POST)

	status := utils.GetContentStatus(param.IsNotice, param.IsSecret)
	result, err := r.db.Exec(
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
	if err != nil {
		return models.FAILED, err
	}

	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED, err
	}
	return uint(insertId), nil
}

// 해시태그와 게시글 번호 연결 정보 저장하기
func (r *TsboardBoardEditRepository) InsertPostHashtag(boardUid uint, postUid uint, hashtagUid uint) error {
	query := fmt.Sprintf("INSERT INTO %s%s (board_uid, post_uid, hashtag_uid) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_POST_HASHTAG)
	_, err := r.db.Exec(query, boardUid, postUid, hashtagUid)
	return err
}

// 신규 태그 저장하기
func (r *TsboardBoardEditRepository) InsertTag(boardUid uint, postUid uint, tag string) (uint, error) {
	query := fmt.Sprintf("INSERT INTO %s%s (name, used, timestamp) VALUES (?, ?, ?)",
		configs.Env.Prefix, models.TABLE_HASHTAG)

	result, err := r.db.Exec(query, tag, 1, time.Now().UnixMilli())
	if err != nil {
		return models.FAILED, err
	}

	hashtagUid, err := result.LastInsertId()
	if err != nil {
		return models.FAILED, err
	}
	return uint(hashtagUid), nil
}

// 게시글에 삽입한 이미지 삭제하기
func (r *TsboardBoardEditRepository) RemoveInsertedImage(imageUid uint, actionUserUid uint) (string, error) {
	query := fmt.Sprintf("SELECT user_uid, path FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_IMAGE)

	var userUid uint
	var path string
	err := r.db.QueryRow(query, imageUid).Scan(&userUid, &path)
	if err != nil {
		return "", err
	}

	if actionUserUid != userUid {
		return "", fmt.Errorf("unauthorized access, only writer can remove an image")
	}

	query = fmt.Sprintf("DELETE FROM %s%s WHERE uid = ? LIMIT 1", configs.Env.Prefix, models.TABLE_IMAGE)
	_, err = r.db.Exec(query, imageUid)
	return path, err
}

// 기존 게시글 수정하기
func (r *TsboardBoardEditRepository) UpdatePost(param models.EditorModifyParameter) error {
	query := fmt.Sprintf(`UPDATE %s%s SET category_uid = ?, title = ?, content = ?, modified = ?, status = ? 
												WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_POST)

	status := utils.GetContentStatus(param.IsNotice, param.IsSecret)
	_, err := r.db.Exec(
		query,
		param.CategoryUid,
		param.Title,
		param.Content,
		time.Now().UnixMilli(),
		status,
		param.PostUid,
	)
	return err
}

// 기존 태그 사용 횟수 올리고 태그와 게시글 번호 연결하기
func (r *TsboardBoardEditRepository) UpdateTag(hashtagUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET used = used + 1 WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_HASHTAG)
	_, err := r.db.Exec(query, hashtagUid)
	return err
}

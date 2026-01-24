package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type CommentRepository interface {
	FindPostUserUidByUid(commentUid uint) (uint, uint)
	GetComments(param models.CommentListParam) ([]models.CommentItem, error)
	GetPostStatus(postUid uint) models.Status
	GetPostWriterUid(postUid uint) uint
	HasReplyComment(commentUid uint) bool
	IsLikedComment(commentUid uint, userUid uint) bool
	InsertComment(param models.CommentWriteParam) (uint, error)
	InsertLikeComment(param models.CommentLikeParam)
	RemoveComment(commentUid uint) error
	UpdateComment(commentUid uint, content string)
	UpdateLikeComment(param models.CommentLikeParam)
	UpdateReplyUid(commentUid uint, replyUid uint)
}

type NuboCommentRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewNuboCommentRepository(db *sql.DB, board BoardRepository) *NuboCommentRepository {
	return &NuboCommentRepository{db: db, board: board}
}

// 댓글 고유 번호로 댓글 작성자의 고유 번호 반환하기
func (r *NuboCommentRepository) FindPostUserUidByUid(commentUid uint) (uint, uint) {
	var postUid, userUid uint
	query := fmt.Sprintf("SELECT post_uid, user_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid).Scan(&postUid, &userUid)
	return postUid, userUid
}

// 게시글 상태 가져오기
func (r *NuboCommentRepository) GetPostStatus(postUid uint) models.Status {
	var status int8
	query := fmt.Sprintf("SELECT status FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, postUid).Scan(&status)
	return models.Status(status)
}

// 게시글 작성자의 고유 번호 반환하기
func (r *NuboCommentRepository) GetPostWriterUid(postUid uint) uint {
	var userUid uint
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, postUid).Scan(&userUid)
	return userUid
}

// 이 댓글에 답글이 하나라도 있는지 확인하기
func (r *NuboCommentRepository) HasReplyComment(commentUid uint) bool {
	var replyUid uint
	query := fmt.Sprintf("SELECT reply_uid FROM %s%s WHERE uid = ? AND status != ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid, models.CONTENT_REMOVED).Scan(&replyUid)
	if replyUid != commentUid {
		return false
	}

	var uid uint
	query = fmt.Sprintf("SELECT uid FROM %s%s WHERE reply_uid = ? AND uid != ? AND status != ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid, commentUid, models.CONTENT_REMOVED).Scan(&uid)
	return uid > 0
}

// 이미 이 댓글에 좋아요를 클릭한 적이 있는지 확인하기
func (r *NuboCommentRepository) IsLikedComment(commentUid uint, userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT comment_uid FROM %s%s WHERE comment_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.QueryRow(query, commentUid, userUid).Scan(&uid)
	return uid > 0
}

// 새로운 댓글 작성하기
func (r *NuboCommentRepository) InsertComment(param models.CommentWriteParam) (uint, error) {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(reply_uid, board_uid, post_uid, user_uid, content, submitted, modified, status) 
												VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_COMMENT)

	result, err := r.db.Exec(
		query,
		0,
		param.BoardUid,
		param.PostUid,
		param.UserUid,
		param.Content,
		time.Now().UnixMilli(),
		0,
		models.CONTENT_NORMAL,
	)
	if err != nil {
		return models.FAILED, err
	}
	insertId, _ := result.LastInsertId()
	return uint(insertId), nil
}

// 이 댓글에 대한 좋아요 추가하기
func (r *NuboCommentRepository) InsertLikeComment(param models.CommentLikeParam) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, comment_uid, user_uid, liked, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.Exec(query, param.BoardUid, param.CommentUid, param.UserUid, param.Liked, time.Now().UnixMilli())
}

// 댓글을 삭제 상태로 변경하기
func (r *NuboCommentRepository) RemoveComment(commentUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)
	_, err := r.db.Exec(query, models.CONTENT_REMOVED, commentUid)
	return err
}

// 기존 댓글 수정하기
func (r *NuboCommentRepository) UpdateComment(commentUid uint, content string) {
	query := fmt.Sprintf("UPDATE %s%s SET content = ?, modified = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.Exec(query, content, time.Now().UnixMilli(), commentUid)
}

// 이 댓글에 대한 좋아요 변경하기
func (r *NuboCommentRepository) UpdateLikeComment(param models.CommentLikeParam) {
	query := fmt.Sprintf("UPDATE %s%s SET liked = ?, timestamp = ? WHERE comment_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.Exec(query, param.Liked, time.Now().UnixMilli(), param.CommentUid, param.UserUid)
}

// 답글 고유 번호 업데이트
func (r *NuboCommentRepository) UpdateReplyUid(commentUid uint, replyUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET reply_uid = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.Exec(query, replyUid, commentUid)
}

// 댓글 목록 가져오기
func (r *NuboCommentRepository) GetComments(param models.CommentListParam) ([]models.CommentItem, error) {
	items := make([]models.CommentItem, 0)
	offset := (param.Page - 1) * param.Limit
	prefix := configs.Env.Prefix

	query := fmt.Sprintf(`SELECT 
			c.uid, c.reply_uid, c.user_uid, c.content, c.submitted, c.modified, c.status,
			u.name, u.profile,
			(SELECT COUNT(*) FROM %s%s WHERE comment_uid = c.uid AND liked = 1),
			EXISTS(SELECT 1 FROM %s%s WHERE comment_uid = c.uid AND user_uid = ? AND liked = 1)
		FROM %s%s AS c
		JOIN (
			SELECT uid FROM %s%s 
			WHERE post_uid = ? AND status IN (?, ?) 
			ORDER BY reply_uid ASC, uid ASC 
			LIMIT ? OFFSET ?
		) AS p ON c.uid = p.uid
		LEFT JOIN %s%s AS u ON c.user_uid = u.uid
		ORDER BY c.reply_uid ASC, c.uid ASC`,
		prefix, models.TABLE_COMMENT_LIKE,
		prefix, models.TABLE_COMMENT_LIKE,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_USER,
	)

	rows, err := r.db.Query(query,
		param.UserUid,
		param.PostUid,
		models.CONTENT_NORMAL,
		models.CONTENT_SECRET,
		param.Limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.CommentItem{}
		err := rows.Scan(
			&item.Uid, &item.ReplyUid, &item.Writer.UserUid, &item.Content, &item.Submitted, &item.Modified, &item.Status,
			&item.Writer.Name, &item.Writer.Profile,
			&item.Like,
			&item.Liked,
		)
		if err == nil {
			item.PostUid = param.PostUid
			items = append(items, item)
		}
	}

	return items, nil
}

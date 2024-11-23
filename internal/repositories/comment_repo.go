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
	GetComments(param models.CommentListParameter) ([]models.CommentItem, error)
	GetLikedCount(commentUid uint) uint
	GetPostStatus(postUid uint) models.Status
	GetMaxUid() uint
	IsLikedComment(commentUid uint, userUid uint) bool
	InsertLikeComment(param models.CommentLikeParameter)
	UpdateLikeComment(param models.CommentLikeParameter)
}

type TsboardCommentRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardCommentRepository(db *sql.DB, board BoardRepository) *TsboardCommentRepository {
	return &TsboardCommentRepository{db: db, board: board}
}

// 댓글 고유 번호로 댓글 작성자의 고유 번호 반환하기
func (r *TsboardCommentRepository) FindPostUserUidByUid(commentUid uint) (uint, uint) {
	var postUid, userUid uint
	query := fmt.Sprintf("SELECT post_uid, user_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.QueryRow(query, commentUid).Scan(&postUid, &userUid)
	return postUid, userUid
}

// 댓글들 가져오기
func (r *TsboardCommentRepository) GetComments(param models.CommentListParameter) ([]models.CommentItem, error) {
	arrow, _ := param.Direction.Query()
	query := fmt.Sprintf(`SELECT uid, reply_uid, user_uid, content, submitted, modified, status 
												FROM %s%s WHERE post_uid = ? AND status != ? AND uid %s ?
												ORDER BY reply_uid ASC LIMIT ?`, configs.Env.Prefix, models.TABLE_COMMENT, arrow)
	rows, err := r.db.Query(query, param.PostUid, models.CONTENT_REMOVED, param.SinceUid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.CommentItem, 0)
	for rows.Next() {
		item := models.CommentItem{}
		err = rows.Scan(&item.Uid, &item.ReplyUid, &item.Writer.UserUid, &item.Content, &item.Submitted, &item.Modified, &item.Status)
		if err != nil {
			return nil, err
		}
		item.PostUid = param.PostUid
		item.Writer = r.board.GetWriterInfo(item.Writer.UserUid)
		item.Like = r.GetLikedCount(item.Uid)
		item.Liked = r.board.CheckLikedComment(item.Uid, param.UserUid)
		items = append(items, item)
	}
	return items, nil
}

// 댓글에 대한 좋아요 수 반환
func (r *TsboardCommentRepository) GetLikedCount(commentUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE comment_uid = ? AND liked = ?",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	r.db.QueryRow(query, commentUid, 1).Scan(&count)
	return count
}

// 게시글 상태 가져오기
func (r *TsboardCommentRepository) GetPostStatus(postUid uint) models.Status {
	var status int8
	query := fmt.Sprintf("SELECT status FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, postUid).Scan(&status)
	return models.Status(status)
}

// 가장 마지막 댓글 고유 번호 가져오기
func (r *TsboardCommentRepository) GetMaxUid() uint {
	var uid uint
	query := fmt.Sprintf("SELECT MAX(uid) FROM %s%s", configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.QueryRow(query).Scan(&uid)
	return uid
}

// 이미 이 댓글에 좋아요를 클릭한 적이 있는지 확인하기
func (r *TsboardCommentRepository) IsLikedComment(commentUid uint, userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT comment_uid FROM %s%s WHERE comment_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	r.db.QueryRow(query, commentUid, userUid).Scan(&uid)
	return uid > 0
}

// 이 댓글에 대한 좋아요 추가하기
func (r *TsboardCommentRepository) InsertLikeComment(param models.CommentLikeParameter) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, comment_uid, user_uid, liked, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	r.db.Exec(query, param.BoardUid, param.CommentUid, param.UserUid, param.Liked, time.Now().UnixMilli())
}

// 이 댓글에 대한 좋아요 변경하기
func (r *TsboardCommentRepository) UpdateLikeComment(param models.CommentLikeParameter) {
	query := fmt.Sprintf("UPDATE %s%s SET liked = ?, timestamp = ? WHERE comment_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	r.db.Exec(query, param.Liked, time.Now().UnixMilli(), param.CommentUid, param.UserUid)
}

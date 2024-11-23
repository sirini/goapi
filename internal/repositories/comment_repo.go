package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type CommentRepository interface {
	GetComments(param models.CommentListParameter) ([]models.CommentItem, error)
	GetPostStatus(postUid uint) models.Status
	GetMaxUid() uint
}

type TsboardCommentRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardCommentRepository(db *sql.DB, board BoardRepository) *TsboardCommentRepository {
	return &TsboardCommentRepository{db: db, board: board}
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
		item.Like = r.board.GetCountByTable(models.TABLE_COMMENT_LIKE, param.PostUid)
		item.Liked = r.board.CheckLikedComment(item.Uid, param.UserUid)
		items = append(items, item)
	}
	return items, nil
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

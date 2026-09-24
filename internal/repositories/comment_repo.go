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
	GetCommentStatus(commentUid uint) models.Status
	GetPostStatus(postUid uint) models.Status
	GetPostWriterUid(postUid uint) uint
	HasReplyComment(commentUid uint) bool
	GetCommentThreadInfo(commentUid uint) models.CommentThreadInfo
	IsLikedComment(commentUid uint, userUid uint) bool
	GetCommentReactionState(commentUid uint, userUid uint) models.ReactionState
	GetCommentUserReaction(commentUid uint, userUid uint) models.ReactionType
	SetCommentReaction(param models.CommentReactionParam) (bool, error)
	IsCommentInBoard(commentUid uint, boardUid uint) bool
	IsCommentInPost(commentUid uint, postUid uint, boardUid uint) bool
	InsertComment(param models.CommentWriteParam, replyUid uint, parentUid uint, depth uint, point models.UpdatePointParam) (uint, error)
	InsertLikeComment(param models.CommentLikeParam)
	RemoveComment(commentUid uint) error
	UpdateComment(commentUid uint, content string)
	UpdateLikeComment(param models.CommentLikeParam)
}

func (r *NuboCommentRepository) IsCommentInBoard(commentUid uint, boardUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE uid = ? AND board_uid = ?)",
		configs.Env.Prefix, models.TABLE_COMMENT)
	if err := r.db.QueryRow(query, commentUid, boardUid).Scan(&exists); err != nil {
		return false
	}
	return exists
}

func (r *NuboCommentRepository) IsCommentInPost(commentUid uint, postUid uint, boardUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE uid = ? AND post_uid = ? AND board_uid = ?)",
		configs.Env.Prefix, models.TABLE_COMMENT)
	if err := r.db.QueryRow(query, commentUid, postUid, boardUid).Scan(&exists); err != nil {
		return false
	}
	return exists
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

// 댓글 상태 가져오기. 삭제된 댓글(답글 자리만 남은 경우 포함)도 이 값으로 판별한다.
func (r *NuboCommentRepository) GetCommentStatus(commentUid uint) models.Status {
	var status int8
	query := fmt.Sprintf("SELECT status FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid).Scan(&status)
	return models.Status(status)
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

// 이 댓글에 직계 자식(답글)이 하나라도 있는지 확인한다. parent_uid 백필 후에는 직계 자식만 보아도 충분하다.
func (r *NuboCommentRepository) HasReplyComment(commentUid uint) bool {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s WHERE parent_uid = ? AND uid != ? AND status != ?)",
		configs.Env.Prefix, models.TABLE_COMMENT)
	if err := r.db.QueryRow(query, commentUid, commentUid, models.CONTENT_REMOVED).Scan(&exists); err != nil {
		return false
	}
	return exists
}

// 댓글의 스레드 정보(스레드 루트 uid·깊이)를 가져온다.
func (r *NuboCommentRepository) GetCommentThreadInfo(commentUid uint) models.CommentThreadInfo {
	var info models.CommentThreadInfo
	query := fmt.Sprintf("SELECT reply_uid, depth FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.QueryRow(query, commentUid).Scan(&info.ReplyUid, &info.Depth)
	return info
}

// 이미 이 댓글에 좋아요를 클릭한 적이 있는지 확인하기
func (r *NuboCommentRepository) IsLikedComment(commentUid uint, userUid uint) bool {
	return r.board.GetCommentUserReaction(commentUid, userUid) == models.REACTION_LIKE
}

// 댓글 종류별 리액션 상태 가져오기
func (r *NuboCommentRepository) GetCommentUserReaction(commentUid uint, userUid uint) models.ReactionType {
	return r.board.GetCommentUserReaction(commentUid, userUid)
}

func (r *NuboCommentRepository) GetCommentReactionState(commentUid uint, userUid uint) models.ReactionState {
	counts := r.board.GetCommentReactionCounts(commentUid)
	current := r.board.GetCommentUserReaction(commentUid, userUid)
	return reactionState(counts, current)
}

// 사용자당 한 행을 유지하며 liked와 reaction_type를 한 트랜잭션으로 동기화한다.
func (r *NuboCommentRepository) SetCommentReaction(param models.CommentReactionParam) (bool, error) {
	code := uint8(param.ReactionCode)
	liked := 0
	if code == 1 {
		liked = 1
	}
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(fmt.Sprintf(`INSERT INTO %s%s (board_uid, comment_uid, user_uid, liked, reaction_type, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE liked = VALUES(liked), reaction_type = VALUES(reaction_type), timestamp = IF(reaction_type <> VALUES(reaction_type) OR liked <> VALUES(liked), VALUES(timestamp), timestamp)`,
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE),
		param.BoardUid, param.CommentUid, param.UserUid, liked, code, time.Now().UnixMilli())
	if err != nil {
		return false, err
	}
	// affected: 0 = 무변경(같은 상태 재설정), 1 = 새 행, 2 = 기존 행 변경
	affected, _ := result.RowsAffected()
	changed := affected != 0
	return changed, tx.Commit()
}

// 새로운 댓글 작성하기
func (r *NuboCommentRepository) InsertComment(param models.CommentWriteParam, replyUid uint, parentUid uint, depth uint, point models.UpdatePointParam) (uint, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.FAILED, err
	}
	defer tx.Rollback()
	if err := applyPointChangeTx(tx, point); err != nil {
		return models.FAILED, err
	}

	query := fmt.Sprintf(`INSERT INTO %s%s 
												(reply_uid, parent_uid, depth, board_uid, post_uid, user_uid, content, submitted, modified, status) 
												VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_COMMENT)

	result, err := tx.Exec(
		query,
		replyUid,
		parentUid,
		depth,
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
	insertId, err := result.LastInsertId()
	if err != nil {
		return models.FAILED, err
	}
	if replyUid == 0 {
		query = fmt.Sprintf("UPDATE %s%s SET reply_uid = ? WHERE uid = ? LIMIT 1",
			configs.Env.Prefix, models.TABLE_COMMENT)
		if _, err := tx.Exec(query, insertId, insertId); err != nil {
			return models.FAILED, err
		}
	}
	if err := tx.Commit(); err != nil {
		return models.FAILED, err
	}
	return uint(insertId), nil
}

// 이 댓글에 대한 좋아요 추가하기
func (r *NuboCommentRepository) InsertLikeComment(param models.CommentLikeParam) {
	code := models.REACTION_NONE
	if param.Liked {
		code = models.REACTION_LIKE
	}
	_, _ = r.SetCommentReaction(models.CommentReactionParam{
		BoardUid: param.BoardUid, CommentUid: param.CommentUid, UserUid: param.UserUid,
		Reaction: code.APIValue(), ReactionCode: code,
	})
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
	r.InsertLikeComment(param)
}

// 댓글 목록 가져오기
func (r *NuboCommentRepository) GetComments(param models.CommentListParam) ([]models.CommentItem, error) {
	items := make([]models.CommentItem, 0)
	userReactionCode := uint8(0)
	offset := (param.Page - 1) * param.Limit
	prefix := configs.Env.Prefix

	query := fmt.Sprintf(`SELECT 
			c.uid, c.reply_uid, c.parent_uid, c.depth, c.user_uid, c.content, c.submitted, c.modified, c.status,
			u.name, u.profile,
			(SELECT COUNT(*) FROM %s%s WHERE comment_uid = c.uid AND liked = 1),
			EXISTS(SELECT 1 FROM %s%s WHERE comment_uid = c.uid AND user_uid = ? AND liked = 1),
			%s,
			COALESCE((SELECT reaction_type FROM %s%s WHERE comment_uid = c.uid AND user_uid = ?), 0)
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
		reactionCountSubselects(models.TABLE_COMMENT_LIKE, "comment_uid", "c.uid"),
		prefix, models.TABLE_COMMENT_LIKE,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_COMMENT,
		prefix, models.TABLE_USER,
	)

	rows, err := r.db.Query(query,
		param.UserUid,
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
		var reactions models.ReactionCountsDTO
		item := models.CommentItem{}
		err := rows.Scan(
			&item.Uid, &item.ReplyUid, &item.ParentUid, &item.Depth, &item.Writer.UserUid, &item.Content, &item.Submitted, &item.Modified, &item.Status,
			&item.Writer.Name, &item.Writer.Profile,
			&item.Like,
			&item.Liked,
			&reactions.Like, &reactions.Best, &reactions.Facepalm, &reactions.Hmm, &reactions.Laugh, &reactions.Celebrate, &reactions.Fire, &reactions.Support, &reactions.Sad, &reactions.Eyes,
			&userReactionCode,
		)
		if err != nil {
			return nil, err
		}
		{
			item.Reactions = reactions
			if reaction := models.ReactionType(userReactionCode); reaction != models.REACTION_NONE {
				value := reaction.APIValue()
				item.MyReaction = &value
			}
			item.PostUid = param.PostUid
			items = append(items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

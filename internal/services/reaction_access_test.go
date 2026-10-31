package services

import (
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

// 리액션 쓰기는 게시판 열람 자격(레벨·포인트)과 삭제된 댓글도 거부해야 한다.
type reactionAccessBoardViewRepo struct {
	repositories.BoardViewRepository
	postBoard map[uint]uint
	banned    map[uint]bool
	needLevel int
	needPoint int
	writerOf  map[uint]uint
	wrote     bool
}

func (r reactionAccessBoardViewRepo) IsPostInBoard(postUid uint, boardUid uint) bool {
	return r.postBoard[postUid] == boardUid
}

func (r reactionAccessBoardViewRepo) CheckBannedByWriter(postUid uint, _ uint) bool {
	return r.banned[postUid]
}

func (r reactionAccessBoardViewRepo) GetNeededLevelPoint(uint, models.BoardAction) (int, int) {
	return r.needLevel, r.needPoint
}

func (r reactionAccessBoardViewRepo) IsWriter(table models.Table, targetUid uint, userUid uint) bool {
	return table == models.TABLE_POST && r.writerOf[targetUid] == userUid
}

func (r reactionAccessBoardViewRepo) GetPostUserReaction(postUid uint, userUid uint) models.ReactionType {
	return models.REACTION_NONE
}

func (r reactionAccessBoardViewRepo) GetPostReactionState(uint, uint) models.ReactionState {
	r.wrote = true
	return models.ReactionState{}
}

type reactionAccessCommentRepo struct {
	repositories.CommentRepository
	commentBoard  map[uint]uint
	commentStatus map[uint]models.Status
	postStatus    map[uint]models.Status
	postOf        map[uint]uint
}

func (r reactionAccessCommentRepo) IsCommentInBoard(commentUid uint, boardUid uint) bool {
	return r.commentBoard[commentUid] == boardUid
}

func (r reactionAccessCommentRepo) GetCommentStatus(commentUid uint) models.Status {
	return r.commentStatus[commentUid]
}

func (r reactionAccessCommentRepo) GetPostStatus(postUid uint) models.Status {
	return r.postStatus[postUid]
}

func (r reactionAccessCommentRepo) FindPostUserUidByUid(commentUid uint) (uint, uint) {
	return r.postOf[commentUid], 0
}

type reactionAccessUserRepo struct {
	repositories.UserRepository
	level int
	point int
}

func (r reactionAccessUserRepo) GetUserLevelPoint(uint) (int, int) { return r.level, r.point }

func TestPostReactionRejectsViewRestrictions(t *testing.T) {
	view := reactionAccessBoardViewRepo{
		postBoard: map[uint]uint{10: 1},
		needLevel: 5,
		needPoint: -100,
	}
	s := NewNuboBoardService(&repositories.Repository{
		BoardView: view.BoardViewRepository,
		Comment:   reactionAccessCommentRepo{postStatus: map[uint]models.Status{10: models.CONTENT_NORMAL}},
		User:      reactionAccessUserRepo{level: 2, point: 10},
	})

	base := models.BoardReactionParam{BoardUid: 1, PostUid: 10, UserUid: 7}
	if _, err := s.SetPostReaction(base); err == nil {
		t.Fatal("below-level user was allowed to react")
	}
	if _, err := s.SetPostReaction(models.BoardReactionParam{
		BoardUid: 1, PostUid: 10, UserUid: 7,
		Reaction: models.REACTIONS.LIKE, ReactionCode: models.REACTION_LIKE,
	}); err == nil {
		t.Fatal("below-level user was allowed to like")
	}

	// 기존 /like의 liked:false도 접근 검사 전에 조기 성공하지 않는다.
	if err := s.LikeThisPost(models.BoardViewLikeParam{
		BoardViewCommonParam: models.BoardViewCommonParam{BoardUid: 1, PostUid: 10, UserUid: 2},
	}); err == nil {
		t.Fatal("restricted user cancel via /like was accepted")
	}

	// 자격이 충분하면 통과한다(취소할 상태가 없어 무변경으로 끝난다).
	s.repos.BoardView = reactionAccessBoardViewRepo{
		postBoard: map[uint]uint{10: 1},
		needLevel: 5,
		needPoint: -100,
	}.BoardViewRepository
	s.repos.User = reactionAccessUserRepo{level: 6, point: 200}
	cancel := base
	cancel.ReactionIsNull = true
	if _, err := s.SetPostReaction(cancel); err != nil {
		t.Fatalf("qualified user was rejected: %v", err)
	}
}

func TestPostReactionRejectsWriterBan(t *testing.T) {
	s := NewNuboBoardService(&repositories.Repository{
		BoardView: reactionAccessBoardViewRepo{
			postBoard: map[uint]uint{10: 1},
			banned:    map[uint]bool{10: true},
		}.BoardViewRepository,
		Comment: reactionAccessCommentRepo{postStatus: map[uint]models.Status{10: models.CONTENT_NORMAL}},
		User:    reactionAccessUserRepo{level: 9, point: 0},
	})
	if _, err := s.SetPostReaction(models.BoardReactionParam{BoardUid: 1, PostUid: 10, UserUid: 7}); err == nil {
		t.Fatal("banned user was allowed to react")
	}
}

func TestCommentReactionRejectsRemovedCommentAndParentPost(t *testing.T) {
	newService := func(commentStatus models.Status, postStatus models.Status) *NuboCommentService {
		return NewNuboCommentService(&repositories.Repository{
			BoardView: reactionAccessBoardViewRepo{postBoard: map[uint]uint{10: 1}}.BoardViewRepository,
			Comment: reactionAccessCommentRepo{
				commentBoard:  map[uint]uint{30: 1},
				commentStatus: map[uint]models.Status{30: commentStatus},
				postStatus:    map[uint]models.Status{10: postStatus},
				postOf:        map[uint]uint{30: 10},
			},
			User: reactionAccessUserRepo{level: 9, point: 0},
		})
	}

	// 삭제된 댓글(답글 자리만 남은 경우 포함)은 새 경로와 기존 /like 모두 거부한다.
	removed := newService(models.CONTENT_REMOVED, models.CONTENT_NORMAL)
	for _, call := range []func() error{
		func() error {
			_, err := removed.SetReaction(models.CommentReactionParam{BoardUid: 1, CommentUid: 30, UserUid: 7})
			return err
		},
		func() error {
			return removed.Like(models.CommentLikeParam{BoardUid: 1, CommentUid: 30, UserUid: 7})
		},
	} {
		if err := call(); err == nil {
			t.Fatal("removed comment was allowed to react")
		}
	}

	// 삭제된 부모 게시글도 거부한다.
	orphan := newService(models.CONTENT_NORMAL, models.CONTENT_REMOVED)
	if _, err := orphan.SetReaction(models.CommentReactionParam{BoardUid: 1, CommentUid: 30, UserUid: 7}); err == nil {
		t.Fatal("comment on removed post was allowed to react")
	}

	// 열람 레벨이 부족하면 거부한다.
	restricted := NewNuboCommentService(&repositories.Repository{
		BoardView: reactionAccessBoardViewRepo{postBoard: map[uint]uint{10: 1}, needLevel: 7}.BoardViewRepository,
		Comment: reactionAccessCommentRepo{
			commentBoard:  map[uint]uint{30: 1},
			commentStatus: map[uint]models.Status{30: models.CONTENT_NORMAL},
			postStatus:    map[uint]models.Status{10: models.CONTENT_NORMAL},
			postOf:        map[uint]uint{30: 10},
		},
		User: reactionAccessUserRepo{level: 2, point: 0},
	})
	if _, err := restricted.SetReaction(models.CommentReactionParam{BoardUid: 1, CommentUid: 30, UserUid: 7}); err == nil {
		t.Fatal("below-level user was allowed to react to comment")
	}
	if err := restricted.Like(models.CommentLikeParam{BoardUid: 1, CommentUid: 30, UserUid: 7}); err == nil {
		t.Fatal("below-level user was allowed to like comment")
	}
}

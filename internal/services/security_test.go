package services

import (
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type ownershipBoardViewRepo struct {
	repositories.BoardViewRepository
	postBoard map[uint]uint
	fileBoard map[uint]uint
	filePost  map[uint]uint
	writers   map[uint]uint
}

func (r ownershipBoardViewRepo) GetFileOwnership(fileUid uint) (uint, uint) {
	return r.fileBoard[fileUid], r.filePost[fileUid]
}

func (r ownershipBoardViewRepo) IsPostInBoard(postUid uint, boardUid uint) bool {
	return r.postBoard[postUid] == boardUid
}

func (r ownershipBoardViewRepo) IsFileInBoard(fileUid uint, boardUid uint) bool {
	return r.fileBoard[fileUid] == boardUid
}

func (r ownershipBoardViewRepo) IsFileInPost(fileUid uint, postUid uint, boardUid uint) bool {
	return r.fileBoard[fileUid] == boardUid && r.filePost[fileUid] == postUid
}

func (r ownershipBoardViewRepo) GetFilePostUid(fileUid uint, _ uint) uint {
	return r.filePost[fileUid]
}

func (r ownershipBoardViewRepo) IsWriter(table models.Table, targetUid uint, userUid uint) bool {
	return table == models.TABLE_POST && r.writers[targetUid] == userUid
}

type postStatusCommentRepo struct {
	repositories.CommentRepository
	statuses map[uint]models.Status
}

func (r postStatusCommentRepo) GetPostStatus(postUid uint) models.Status {
	return r.statuses[postUid]
}

type ownershipCommentRepo struct {
	repositories.CommentRepository
	commentBoard map[uint]uint
	commentPost  map[uint]uint
}

func (r ownershipCommentRepo) IsCommentInBoard(commentUid uint, boardUid uint) bool {
	return r.commentBoard[commentUid] == boardUid
}

func (r ownershipCommentRepo) IsCommentInPost(commentUid uint, postUid uint, boardUid uint) bool {
	return r.commentBoard[commentUid] == boardUid && r.commentPost[commentUid] == postUid
}

type denyAuthRepo struct{ repositories.AuthRepository }

func (denyAuthRepo) CheckPermissionByUid(uint, uint) bool { return false }

type boardPermissionRepo struct {
	repositories.AuthRepository
	permissions map[uint]bool
}

func (r boardPermissionRepo) CheckPermissionByUid(_ uint, boardUid uint) bool {
	return r.permissions[boardUid]
}

type boardConfigRepo struct {
	repositories.BoardRepository
	configs map[uint]models.BoardConfig
}

func (r boardConfigRepo) GetBoardConfig(boardUid uint) models.BoardConfig {
	return r.configs[boardUid]
}

type postMoveBoardViewRepo struct {
	repositories.BoardViewRepository
	postBoard map[uint]uint
	boards    []models.BoardItem
	exists    map[uint]bool
	movedTo   uint
	movedPost uint
}

func (r *postMoveBoardViewRepo) IsPostInBoard(postUid uint, boardUid uint) bool {
	return r.postBoard[postUid] == boardUid
}

func (r *postMoveBoardViewRepo) GetAllBoards() []models.BoardItem { return r.boards }

func (r *postMoveBoardViewRepo) BoardExists(boardUid uint) bool { return r.exists[boardUid] }

func (r *postMoveBoardViewRepo) MovePost(boardUid uint, postUid uint) error {
	r.movedTo = boardUid
	r.movedPost = postUid
	return nil
}

func TestBoardMoveTargetsOnlyIncludeBoardsUserCanManage(t *testing.T) {
	boardView := &postMoveBoardViewRepo{boards: []models.BoardItem{
		{Pair: models.Pair{Uid: 1, Name: "Source"}, Id: "source", Type: models.BOARD_BOARD},
		{Pair: models.Pair{Uid: 2, Name: "Gallery"}, Id: "gallery", Type: models.BOARD_GALLERY},
		{Pair: models.Pair{Uid: 3, Name: "Other"}, Id: "other", Type: models.BOARD_BOARD},
		{Pair: models.Pair{Uid: 4, Name: "Market"}, Id: "market", Type: models.BOARD_TRADE},
	}}
	s := NewNuboBoardService(&repositories.Repository{
		Auth:      boardPermissionRepo{permissions: map[uint]bool{1: true, 2: true, 4: true}},
		Board:     boardConfigRepo{configs: map[uint]models.BoardConfig{1: {Type: models.BOARD_BOARD}}},
		BoardView: boardView,
	})

	targets, err := s.GetBoardList(1, 7)
	if err != nil {
		t.Fatalf("GetBoardList returned an error: %v", err)
	}
	if len(targets) != 1 || targets[0].Uid != 2 || targets[0].Id != "gallery" {
		t.Fatalf("unexpected move targets: %+v", targets)
	}

	s.repos.Auth = denyAuthRepo{}
	if _, err := s.GetBoardList(1, 7); err == nil {
		t.Fatal("non-admin received move targets")
	}
}

func TestMovePostRequiresPermissionForSourceAndTarget(t *testing.T) {
	boardView := &postMoveBoardViewRepo{
		postBoard: map[uint]uint{10: 1},
		exists:    map[uint]bool{2: true},
	}
	auth := boardPermissionRepo{permissions: map[uint]bool{1: true, 2: false}}
	s := NewNuboBoardService(&repositories.Repository{
		Auth: auth,
		Board: boardConfigRepo{configs: map[uint]models.BoardConfig{
			1: {Type: models.BOARD_BOARD},
			2: {Type: models.BOARD_GALLERY},
		}},
		BoardView: boardView,
	})
	param := models.BoardMovePostParam{
		BoardViewCommonParam: models.BoardViewCommonParam{BoardUid: 1, PostUid: 10, UserUid: 7},
		TargetBoardUid:       2,
	}

	if err := s.MovePost(param); err == nil {
		t.Fatal("move into a board the user cannot manage was accepted")
	}
	if boardView.movedPost != 0 {
		t.Fatal("repository move was called after authorization failed")
	}

	auth.permissions[2] = true
	if err := s.MovePost(param); err != nil {
		t.Fatalf("authorized move failed: %v", err)
	}
	if boardView.movedTo != 2 || boardView.movedPost != 10 {
		t.Fatalf("unexpected repository move: board=%d post=%d", boardView.movedTo, boardView.movedPost)
	}
}

func TestBoardOperationsRejectCrossBoardIdentifiers(t *testing.T) {
	boardView := ownershipBoardViewRepo{
		postBoard: map[uint]uint{10: 1},
		fileBoard: map[uint]uint{20: 1},
	}
	s := NewNuboBoardService(&repositories.Repository{BoardView: boardView})

	if _, err := s.GetViewItem(models.BoardViewParam{BoardUid: 2, PostUid: 10}); err == nil {
		t.Fatal("cross-board post view was accepted")
	}
	if _, err := s.Download(2, 20, 3); err == nil {
		t.Fatal("cross-board file download was accepted")
	}
	if err := s.LikeThisPost(models.BoardViewLikeParam{BoardViewCommonParam: models.BoardViewCommonParam{BoardUid: 2, PostUid: 10}}); err == nil {
		t.Fatal("cross-board post like was accepted")
	}
}

func TestBoardMutationsRejectCrossBoardIdentifiers(t *testing.T) {
	boardView := ownershipBoardViewRepo{
		postBoard: map[uint]uint{10: 1},
		fileBoard: map[uint]uint{20: 1},
		filePost:  map[uint]uint{20: 10},
	}
	s := NewNuboBoardService(&repositories.Repository{BoardView: boardView})

	if _, err := s.LoadPost(2, 10, 7); err == nil {
		t.Fatal("cross-board post edit load was accepted")
	}
	if err := s.ModifyPost(models.EditorModifyParam{
		EditorWriteParam: models.EditorWriteParam{BoardUid: 2, UserUid: 7},
		PostUid:          10,
	}); err == nil {
		t.Fatal("cross-board post modification was accepted")
	}
	if err := s.RemovePost(2, 10, 7); err == nil {
		t.Fatal("cross-board post removal was accepted")
	}
	if err := s.MovePost(models.BoardMovePostParam{
		BoardViewCommonParam: models.BoardViewCommonParam{BoardUid: 2, PostUid: 10, UserUid: 7},
		TargetBoardUid:       3,
	}); err == nil {
		t.Fatal("cross-board post move was accepted")
	}
	if err := s.RemoveAttachedFile(models.EditorRemoveAttachedParam{
		BoardUid: 2,
		PostUid:  10,
		FileUid:  20,
		UserUid:  7,
	}); err == nil {
		t.Fatal("cross-board attachment removal was accepted")
	}
}

func TestCommentOperationsRejectCrossBoardIdentifiers(t *testing.T) {
	boardView := ownershipBoardViewRepo{postBoard: map[uint]uint{10: 1}}
	comments := ownershipCommentRepo{
		commentBoard: map[uint]uint{30: 1},
		commentPost:  map[uint]uint{30: 10},
	}
	s := NewNuboCommentService(&repositories.Repository{BoardView: boardView, Comment: comments})

	if _, err := s.List(models.CommentListParam{BoardUid: 2, PostUid: 10, UserUid: 7}); err == nil {
		t.Fatal("cross-board comment list was accepted")
	}
	if _, err := s.Write(models.CommentWriteParam{BoardUid: 2, PostUid: 10, UserUid: 7}); err == nil {
		t.Fatal("cross-board comment write was accepted")
	}
	if err := s.Like(models.CommentLikeParam{BoardUid: 2, CommentUid: 30, UserUid: 7}); err == nil {
		t.Fatal("cross-board comment like was accepted")
	}
	if err := s.Remove(models.CommentRemoveParam{BoardUid: 2, RemoveTargetUid: 30, UserUid: 7}); err == nil {
		t.Fatal("cross-board comment removal was accepted")
	}
	write := models.CommentWriteParam{BoardUid: 2, PostUid: 10, UserUid: 7}
	if err := s.Modify(models.CommentModifyParam{CommentWriteParam: write, ModifyTargetUid: 30}); err == nil {
		t.Fatal("cross-board comment modification was accepted")
	}
	if _, err := s.Reply(models.CommentReplyParam{CommentWriteParam: write, ReplyTargetUid: 30}); err == nil {
		t.Fatal("cross-board comment reply was accepted")
	}
}

func TestDownloadRejectsSecretPostAttachmentForNonOwner(t *testing.T) {
	boardView := ownershipBoardViewRepo{
		fileBoard: map[uint]uint{20: 1},
		filePost:  map[uint]uint{20: 10},
		writers:   map[uint]uint{10: 7},
	}
	repos := &repositories.Repository{
		Auth:      denyAuthRepo{},
		BoardView: boardView,
		Comment:   postStatusCommentRepo{statuses: map[uint]models.Status{10: models.CONTENT_SECRET}},
	}
	s := NewNuboBoardService(repos)
	if _, err := s.Download(1, 20, 8); err == nil {
		t.Fatal("secret-post attachment was downloadable by a non-owner")
	}
}

func TestThumbnailRejectsNonOwner(t *testing.T) {
	boardView := ownershipBoardViewRepo{
		fileBoard: map[uint]uint{20: 1},
		filePost:  map[uint]uint{20: 10},
		writers:   map[uint]uint{10: 7},
	}
	s := NewNuboBoardService(&repositories.Repository{Auth: denyAuthRepo{}, BoardView: boardView})
	if _, err := s.GetThumbnailImage(20, 8); err == nil {
		t.Fatal("attachment preview was returned to a non-owner")
	}
}

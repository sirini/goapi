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

type denyAuthRepo struct{ repositories.AuthRepository }

func (denyAuthRepo) CheckPermissionByUid(uint, uint) bool { return false }

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

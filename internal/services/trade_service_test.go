package services

import (
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type tradeBoardRepo struct{ repositories.BoardRepository }

func (tradeBoardRepo) GetBoardConfig(boardUid uint) models.BoardConfig {
	return models.BoardConfig{Uid: boardUid, Type: models.BOARD_TRADE}
}

type tradeAuthRepo struct {
	repositories.AuthRepository
	admins map[uint]bool
}

func (r tradeAuthRepo) CheckPermissionByUid(userUid uint, _ uint) bool { return r.admins[userUid] }

type statusTradeRepo struct {
	repositories.TradeRepository
	updated bool
}

func (r *statusTradeRepo) UpdateStatus(uint, models.TradeStatus) error {
	r.updated = true
	return nil
}

func TestHasProductImage(t *testing.T) {
	if !hasProductImage(`<p><img src="/upload/item.webp"></p>`, nil) {
		t.Fatal("inline product image was not detected")
	}
	image := &multipart.FileHeader{Filename: "item.jpg", Header: textproto.MIMEHeader{}}
	if !hasProductImage("<p>description</p>", []*multipart.FileHeader{image}) {
		t.Fatal("attached product image was not detected")
	}
	if hasProductImage("<p>description</p>", nil) {
		t.Fatal("image-less listing was accepted")
	}
}

func TestMergeTradeItemsPreservesPostAssociation(t *testing.T) {
	posts := []models.BoardListItem{{BoardCommonPostItem: models.BoardCommonPostItem{Uid: 3}}}
	items, err := mergeTradeItems(posts, map[uint]models.TradeResult{3: {Uid: 9}})
	if err != nil || len(items) != 1 || items[0].Post.Uid != 3 || items[0].Trade.Uid != 9 {
		t.Fatalf("unexpected merged items: %#v, %v", items, err)
	}
	if _, err := mergeTradeItems(posts, nil); err == nil {
		t.Fatal("missing trade metadata was silently accepted")
	}
}

func TestTradeStatusRequiresOwnerOrBoardAdmin(t *testing.T) {
	boardView := ownershipBoardViewRepo{postBoard: map[uint]uint{10: 2}, writers: map[uint]uint{10: 7}}
	tradeRepo := &statusTradeRepo{}
	repos := &repositories.Repository{
		Auth:      tradeAuthRepo{admins: map[uint]bool{8: true}},
		Board:     tradeBoardRepo{},
		BoardView: boardView,
		Trade:     tradeRepo,
	}
	service := NewNuboTradeService(repos, nil)

	base := models.TradeStatusParam{BoardUid: 2, PostUid: 10, Status: models.TRADE_SOLD}
	base.UserUid = 9
	if err := service.UpdateTradeStatus(base); err == nil {
		t.Fatal("unrelated user changed the trade status")
	}
	base.UserUid = 8
	if err := service.UpdateTradeStatus(base); err != nil || !tradeRepo.updated {
		t.Fatalf("board administrator could not change status: %v", err)
	}
}

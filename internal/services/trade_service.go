package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type TradeService interface {
	GetTradeItem(postUid uint, userUid uint) (models.TradeResult, error)
	ModifyPost(param models.TradeWriterParam) error
	UpdateStatus(postUid uint, newStatus uint, userUid uint) error
	WritePost(param models.TradeWriterParam) error
}

type NuboTradeService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboTradeService(repos *repositories.Repository) *NuboTradeService {
	return &NuboTradeService{repos: repos}
}

// 물품 거래 보기
func (s *NuboTradeService) GetTradeItem(postUid uint, userUid uint) (models.TradeResult, error) {
	return s.repos.Trade.GetTradeItem(postUid)
}

// 물품 거래 수정하기
func (s *NuboTradeService) ModifyPost(param models.TradeWriterParam) error {
	if isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid); !isWriter {
		return fmt.Errorf("only the author of the post can modify")
	}
	return s.repos.Trade.UpdateTrade(param)
}

// 거래 상태 변경하기
func (s *NuboTradeService) UpdateStatus(postUid uint, newStatus uint, userUid uint) error {
	if isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid); !isWriter {
		return fmt.Errorf("only the author of the post can change the transaction status")
	}
	return s.repos.Trade.UpdateStatus(postUid, newStatus)
}

// 물품 거래 게시글 작성하기
func (s *NuboTradeService) WritePost(param models.TradeWriterParam) error {
	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return fmt.Errorf("you have no permission to write a new trade post")
	}
	return s.repos.Trade.InsertTrade(param)
}

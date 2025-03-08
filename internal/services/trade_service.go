package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type TradeService interface {
	GetTradeItem(postUid uint, userUid uint) (models.TradeResult, error)
	ModifyPost(param models.TradeWriterParameter) error
	UpdateStatus(postUid uint, newStatus uint, userUid uint) error
	WritePost(param models.TradeWriterParameter) error
}

type TsboardTradeService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardTradeService(repos *repositories.Repository) *TsboardTradeService {
	return &TsboardTradeService{repos: repos}
}

// 물품 거래 보기
func (s *TsboardTradeService) GetTradeItem(postUid uint, userUid uint) (models.TradeResult, error) {
	return s.repos.Trade.GetTradeItem(postUid)
}

// 물품 거래 수정하기
func (s *TsboardTradeService) ModifyPost(param models.TradeWriterParameter) error {
	if isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid); !isWriter {
		return fmt.Errorf("only the author of the post can modify")
	}
	return s.repos.Trade.UpdateTrade(param)
}

// 거래 상태 변경하기
func (s *TsboardTradeService) UpdateStatus(postUid uint, newStatus uint, userUid uint) error {
	if isWriter := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid); !isWriter {
		return fmt.Errorf("only the author of the post can change the transaction status")
	}
	return s.repos.Trade.UpdateStatus(postUid, newStatus)
}

// 물품 거래 게시글 작성하기
func (s *TsboardTradeService) WritePost(param models.TradeWriterParameter) error {
	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return fmt.Errorf("you have no permission to write a new trade post")
	}
	return s.repos.Trade.InsertTrade(param)
}

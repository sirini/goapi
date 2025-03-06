package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type TradeService interface {
	GetTradeItem(postUid uint, userUid uint) (models.TradeResult, error)
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

// 물품 거래 게시글 작성하기
func (s *TsboardTradeService) WritePost(param models.TradeWriterParameter) error {
	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return fmt.Errorf("you have no permission to write a new trade post")
	}
	return s.repos.Trade.InsertTrade(param)
}

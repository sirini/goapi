package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
)

type TradeService interface {
	ChangeFavorite(tradeUid uint, userUid uint, isAdded bool) error
}

type TsboardTradeService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardTradeService(repos *repositories.Repository) *TsboardTradeService {
	return &TsboardTradeService{repos: repos}
}

// 찜하기에 추가(취소) 하기
func (s *TsboardTradeService) ChangeFavorite(tradeUid uint, userUid uint, isAdded bool) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

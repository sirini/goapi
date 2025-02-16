package repositories

import (
	"database/sql"
	"fmt"
)

type TradeRepository interface {
	UpdateFavorite(tradeUid uint, userUid uint, isAdded bool) error
}

type TsboardTradeRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardTradeRepository(db *sql.DB) *TsboardTradeRepository {
	return &TsboardTradeRepository{db: db}
}

// 찜하기에 추가(혹은 취소) 하기
func (r *TsboardTradeRepository) UpdateFavorite(tradeUid uint, userUid uint, isAdded bool) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

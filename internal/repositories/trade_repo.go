package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type TradeRepository interface {
	GetTotalFavorite(tradeUid uint) uint
	GetTradeItem(postUid uint) (models.TradeResult, error)
	InsertTrade(param models.TradeWriterParameter) error
	IsFavorited(tradeUid uint, userUid uint) bool
	UpdateFavorite(tradeUid uint, userUid uint, isAdded bool) error
}

type TsboardTradeRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardTradeRepository(db *sql.DB) *TsboardTradeRepository {
	return &TsboardTradeRepository{db: db}
}

// 물품 거래 게시글의 찜 개수 반환
func (r *TsboardTradeRepository) GetTotalFavorite(tradeUid uint) uint {
	var count uint
	query := fmt.Sprintf(`SELECT COUNT(*) AS total_count FROM %s%s WHERE trade_uid = ? AND favorited = ?`,
		configs.Env.Prefix, models.TABLE_TRADE_FAVORITE)
	r.db.QueryRow(query, tradeUid, 1).Scan(&count)
	return count
}

// 물품 거래 내역 가져오기
func (r *TsboardTradeRepository) GetTradeItem(postUid uint) (models.TradeResult, error) {
	item := models.TradeResult{}
	query := fmt.Sprintf(`
		SELECT uid, brand, category, price, product_condition, location, shipping_type, status, completed 
		FROM %s%s WHERE post_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_TRADE_PRODUCT)
	err := r.db.QueryRow(query, postUid).Scan(
		&item.Uid,
		&item.Brand,
		&item.ProductCategory,
		&item.Price,
		&item.ProductCondition,
		&item.Location,
		&item.ShippingType,
		&item.Status,
		&item.Completed,
	)
	return item, err
}

// 새 물품 거래 게시글 등록
func (r *TsboardTradeRepository) InsertTrade(param models.TradeWriterParameter) error {
	query := fmt.Sprintf(`INSERT INTO %s%s
		(post_uid, brand, category, price, product_condition, location, shipping_type, status, completed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_TRADE_PRODUCT)

	_, err := r.db.Exec(
		query,
		param.PostUid,
		param.Brand,
		param.ProductCategory,
		param.Price,
		param.ProductCondition,
		param.Location,
		param.ShippingType,
		param.Status,
		0)
	return err
}

// 사용자가 이 물품 거래 게시글을 찜했는지 확인하기
func (r *TsboardTradeRepository) IsFavorited(tradeUid uint, userUid uint) bool {
	var uid uint
	query := fmt.Sprintf(`SELECT uid FROM %s%s WHERE trade_uid = ? AND user_uid = ? AND favorited = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_TRADE_FAVORITE)
	r.db.QueryRow(query, tradeUid, userUid, 1).Scan(&uid)
	return uid > 0
}

// 찜하기에 추가(혹은 취소) 하기
func (r *TsboardTradeRepository) UpdateFavorite(tradeUid uint, userUid uint, isAdded bool) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

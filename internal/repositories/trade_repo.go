package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type TradeRepository interface {
	GetTradeItem(postUid uint) (models.TradeResult, error)
	InsertTrade(param models.TradeWriterParam) error
	UpdateStatus(postUid uint, newStatus uint) error
	UpdateTrade(param models.TradeWriterParam) error
}

type NuboTradeRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewNuboTradeRepository(db *sql.DB) *NuboTradeRepository {
	return &NuboTradeRepository{db: db}
}

// 물품 거래 내역 가져오기
func (r *NuboTradeRepository) GetTradeItem(postUid uint) (models.TradeResult, error) {
	item := models.TradeResult{}
	query := fmt.Sprintf(`
		SELECT uid, brand, category, price, product_condition, location, shipping_type, status, completed 
		FROM %s%s WHERE post_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_TRADE)
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
func (r *NuboTradeRepository) InsertTrade(param models.TradeWriterParam) error {
	query := fmt.Sprintf(`INSERT INTO %s%s
		(post_uid, brand, category, price, product_condition, location, shipping_type, status, completed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_TRADE)

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

// 거래 상태 업데이트
func (r *NuboTradeRepository) UpdateStatus(postUid uint, newStatus uint) error {
	completed := ""
	if newStatus == models.TRADE_DONE {
		completed = fmt.Sprintf(", completed = %d", time.Now().UnixMilli())
	}

	query := fmt.Sprintf(`UPDATE %s%s SET status = ? %s WHERE post_uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_TRADE, completed)
	_, err := r.db.Exec(query, newStatus, postUid)
	return err
}

// 물품 거래 업데이트
func (r *NuboTradeRepository) UpdateTrade(param models.TradeWriterParam) error {
	query := fmt.Sprintf(`UPDATE %s%s SET brand = ?, category = ?, price = ?, product_condition = ?, location = ?, shipping_type = ? WHERE post_uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_TRADE)
	_, err := r.db.Exec(
		query,
		param.Brand,
		param.ProductCategory,
		param.Price,
		param.ProductCondition,
		param.Location,
		param.ShippingType,
		param.PostUid,
	)
	return err
}

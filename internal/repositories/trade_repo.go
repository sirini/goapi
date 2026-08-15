package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type TradeRepository interface {
	GetTradeItem(postUid uint) (models.TradeResult, error)
	GetTradeItems(postUids []uint) (map[uint]models.TradeResult, error)
	InsertTradePost(param models.TradeWriteParam, point models.UpdatePointParam) (uint, error)
	UpdateStatus(postUid uint, newStatus models.TradeStatus) error
	UpdateTradePost(param models.TradeModifyParam) error
}

type NuboTradeRepository struct {
	db *sql.DB
}

func (r *NuboTradeRepository) GetTradeItems(postUids []uint) (map[uint]models.TradeResult, error) {
	items := make(map[uint]models.TradeResult, len(postUids))
	if len(postUids) == 0 {
		return items, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(postUids)), ",")
	args := make([]any, len(postUids))
	for i, uid := range postUids {
		args[i] = uid
	}
	query := fmt.Sprintf(`SELECT post_uid, uid, brand, price, price_type, currency, product_condition,
		location, shipping_type, status, completed FROM %s%s WHERE post_uid IN (%s)`,
		configs.Env.Prefix, models.TABLE_TRADE, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var postUid uint
		item := models.TradeResult{}
		if err := rows.Scan(&postUid, &item.Uid, &item.Brand, &item.Price, &item.PriceType, &item.Currency,
			&item.ProductCondition, &item.Location, &item.ShippingType, &item.Status, &item.Completed); err != nil {
			return nil, err
		}
		items[postUid] = item
	}
	return items, rows.Err()
}

func (r *NuboTradeRepository) InsertTradePost(param models.TradeWriteParam, point models.UpdatePointParam) (uint, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.FAILED, err
	}
	defer tx.Rollback()
	if err := applyPointChangeTx(tx, point); err != nil {
		return models.FAILED, err
	}
	postQuery := fmt.Sprintf(`INSERT INTO %s%s
		(board_uid, user_uid, category_uid, title, content, submitted, modified, hit, status)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0, ?)`, configs.Env.Prefix, models.TABLE_POST)
	postResult, err := tx.Exec(postQuery, param.BoardUid, param.UserUid, param.CategoryUid, param.Title,
		param.Content, time.Now().UnixMilli(), utils.GetContentStatus(param.IsNotice, param.IsSecret))
	if err != nil {
		return models.FAILED, err
	}
	postUid, err := postResult.LastInsertId()
	if err != nil {
		return models.FAILED, err
	}
	tradeQuery := fmt.Sprintf(`INSERT INTO %s%s
		(post_uid, brand, price, price_type, currency, product_condition, location, shipping_type, status, completed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`, configs.Env.Prefix, models.TABLE_TRADE)
	if _, err := tx.Exec(tradeQuery, postUid, param.Brand, param.Price, param.PriceType, param.Currency,
		param.ProductCondition, param.Location, param.ShippingType, models.TRADE_AVAILABLE); err != nil {
		return models.FAILED, err
	}
	if err := tx.Commit(); err != nil {
		return models.FAILED, err
	}
	return uint(postUid), nil
}

func (r *NuboTradeRepository) UpdateTradePost(param models.TradeModifyParam) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	postQuery := fmt.Sprintf(`UPDATE %s%s SET category_uid = ?, title = ?, content = ?, modified = ?, status = ?
		WHERE uid = ? AND board_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_POST)
	postResult, err := tx.Exec(postQuery, param.CategoryUid, param.Title, param.Content, time.Now().UnixMilli(),
		utils.GetContentStatus(param.IsNotice, param.IsSecret), param.PostUid, param.BoardUid)
	if err != nil {
		return err
	}
	if affected, _ := postResult.RowsAffected(); affected != 1 {
		return fmt.Errorf("post not found")
	}
	tradeQuery := fmt.Sprintf(`UPDATE %s%s SET brand = ?, price = ?, price_type = ?, currency = ?,
		product_condition = ?, location = ?, shipping_type = ? WHERE post_uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_TRADE)
	tradeResult, err := tx.Exec(tradeQuery, param.Brand, param.Price, param.PriceType, param.Currency,
		param.ProductCondition, param.Location, param.ShippingType, param.PostUid)
	if err != nil {
		return err
	}
	if affected, _ := tradeResult.RowsAffected(); affected != 1 {
		return fmt.Errorf("trade metadata not found")
	}
	return tx.Commit()
}

// sql.DB 포인터 주입받기
func NewNuboTradeRepository(db *sql.DB) *NuboTradeRepository {
	return &NuboTradeRepository{db: db}
}

// 물품 거래 내역 가져오기
func (r *NuboTradeRepository) GetTradeItem(postUid uint) (models.TradeResult, error) {
	item := models.TradeResult{}
	query := fmt.Sprintf(`
		SELECT uid, brand, price, price_type, currency, product_condition, location, shipping_type, status, completed
		FROM %s%s WHERE post_uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_TRADE)
	err := r.db.QueryRow(query, postUid).Scan(
		&item.Uid,
		&item.Brand,
		&item.Price,
		&item.PriceType,
		&item.Currency,
		&item.ProductCondition,
		&item.Location,
		&item.ShippingType,
		&item.Status,
		&item.Completed,
	)
	return item, err
}

// 거래 상태 업데이트
func (r *NuboTradeRepository) UpdateStatus(postUid uint, newStatus models.TradeStatus) error {
	completed := ""
	if newStatus == models.TRADE_SOLD {
		completed = fmt.Sprintf(", completed = %d", time.Now().UnixMilli())
	} else {
		completed = ", completed = 0"
	}

	query := fmt.Sprintf(`UPDATE %s%s SET status = ? %s WHERE post_uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_TRADE, completed)
	_, err := r.db.Exec(query, newStatus, postUid)
	return err
}

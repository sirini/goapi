package models

type TradePriceType uint8
type TradeProductCondition uint8
type TradeShippingType uint8
type TradeStatus uint8

const (
	TRADE_PRICE_FIXED TradePriceType = iota
	TRADE_PRICE_NEGOTIABLE
	TRADE_PRICE_FREE
)

const (
	TRADE_CONDITION_UNOPENED TradeProductCondition = iota
	TRADE_CONDITION_LIKE_NEW
	TRADE_CONDITION_USED
	TRADE_CONDITION_WORN
	TRADE_CONDITION_DAMAGED
)

const (
	TRADE_SHIPPING_PARCEL TradeShippingType = iota
	TRADE_SHIPPING_MEETUP
	TRADE_SHIPPING_BOTH
)

const (
	TRADE_AVAILABLE TradeStatus = iota
	TRADE_RESERVED
	TRADE_SOLD
	TRADE_WITHDRAWN
)

func (v TradePriceType) Valid() bool        { return v <= TRADE_PRICE_FREE }
func (v TradeProductCondition) Valid() bool { return v <= TRADE_CONDITION_DAMAGED }
func (v TradeShippingType) Valid() bool     { return v <= TRADE_SHIPPING_BOTH }
func (v TradeStatus) Valid() bool           { return v <= TRADE_WITHDRAWN }

func (v TradeShippingType) NeedsLocation() bool {
	return v == TRADE_SHIPPING_MEETUP || v == TRADE_SHIPPING_BOTH
}

// 물품 거래 공통 항목 정의
type TradeCommonItem struct {
	Brand            string                `json:"brand"`
	Price            uint64                `json:"price"`
	PriceType        TradePriceType        `json:"priceType"`
	Currency         string                `json:"currency"`
	ProductCondition TradeProductCondition `json:"productCondition"`
	Location         string                `json:"location"`
	ShippingType     TradeShippingType     `json:"shippingType"`
	Status           TradeStatus           `json:"status"`
}

// 물품 거래 작성용 파라미터 정의
type TradeWriterParam struct {
	TradeCommonItem
	PostUid uint
	UserUid uint
}

// 물품 거래 내용 정의
type TradeResult struct {
	TradeCommonItem
	Uid       uint   `json:"uid"`
	Completed uint64 `json:"completed"`
}

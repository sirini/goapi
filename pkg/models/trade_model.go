package models

// 물품 거래 공통 항목 정의
type TradeCommonItem struct {
	Brand            string `json:"brand"`
	ProductCategory  uint8  `json:"productCategory"`
	Price            uint   `json:"price"`
	ProductCondition uint8  `json:"productCondition"`
	Location         string `json:"location"`
	ShippingType     uint8  `json:"shippingType"`
	Status           uint8  `json:"status"`
}

// 물품 거래 작성용 파라미터 정의
type TradeWriterParameter struct {
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

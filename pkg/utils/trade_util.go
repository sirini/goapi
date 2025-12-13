package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
)

// 물품 거래 글 작성/수정 시 파라미터 검사 및 타입 변환
func CheckTradeWriteParams(c fiber.Ctx) (models.TradeWriterParam, error) {
	result := models.TradeWriterParam{}
	actionUserUid := ExtractUserUid(c.Get(models.AUTH_KEY))
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return result, err
	}
	brand := Escape(c.FormValue("brand"))
	if len(brand) < 2 {
		return result, fmt.Errorf("invalid brand name, too short")
	}
	productCategory, err := strconv.ParseUint(c.FormValue("productCategory"), 10, 32)
	if err != nil {
		return result, err
	}
	price, err := strconv.ParseUint(c.FormValue("price"), 10, 32)
	if err != nil {
		return result, err
	}
	productCondition, err := strconv.ParseUint(c.FormValue("productCondition"), 10, 32)
	if err != nil {
		return result, err
	}
	location := Escape(c.FormValue("location"))
	if len(location) < 2 {
		return result, fmt.Errorf("invalid location, too short")
	}
	shippingType, err := strconv.ParseUint(c.FormValue("shippingType"), 10, 32)
	if err != nil {
		return result, err
	}
	status, err := strconv.ParseUint(c.FormValue("status"), 10, 32)
	if err != nil {
		return result, err
	}

	result = models.TradeWriterParam{
		PostUid: uint(postUid),
		UserUid: uint(actionUserUid),
		TradeCommonItem: models.TradeCommonItem{
			Brand:            brand,
			ProductCategory:  uint8(productCategory),
			Price:            uint(price),
			ProductCondition: uint8(productCondition),
			Location:         location,
			ShippingType:     uint8(shippingType),
			Status:           uint8(status),
		},
	}
	return result, nil
}

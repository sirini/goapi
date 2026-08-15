package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
)

func CheckTradeParams(c fiber.Ctx) (models.TradeCommonItem, error) {
	result := models.TradeCommonItem{}
	brand := Escape(c.FormValue("brand"))
	if utf8.RuneCountInString(brand) > 100 {
		return result, fmt.Errorf("invalid brand name, too long")
	}
	price, err := strconv.ParseUint(c.FormValue("price"), 10, 64)
	if err != nil {
		return result, fmt.Errorf("invalid price")
	}
	priceTypeValue, err := strconv.ParseUint(c.FormValue("priceType"), 10, 8)
	if err != nil {
		return result, fmt.Errorf("invalid price type")
	}
	priceType := models.TradePriceType(priceTypeValue)
	if !priceType.Valid() {
		return result, fmt.Errorf("invalid price type")
	}
	if priceType == models.TRADE_PRICE_FREE {
		price = 0
	} else if price == 0 {
		return result, fmt.Errorf("price must be greater than zero")
	}
	currency := strings.ToUpper(strings.TrimSpace(c.FormValue("currency")))
	if currency == "" {
		currency = "KRW"
	}
	if len(currency) != 3 || currency[0] < 'A' || currency[0] > 'Z' ||
		currency[1] < 'A' || currency[1] > 'Z' || currency[2] < 'A' || currency[2] > 'Z' {
		return result, fmt.Errorf("invalid currency")
	}
	productConditionValue, err := strconv.ParseUint(c.FormValue("productCondition"), 10, 8)
	if err != nil {
		return result, fmt.Errorf("invalid product condition")
	}
	productCondition := models.TradeProductCondition(productConditionValue)
	if !productCondition.Valid() {
		return result, fmt.Errorf("invalid product condition")
	}
	location := Escape(c.FormValue("location"))
	if utf8.RuneCountInString(location) > 100 {
		return result, fmt.Errorf("invalid location, too long")
	}
	shippingTypeValue, err := strconv.ParseUint(c.FormValue("shippingType"), 10, 8)
	if err != nil {
		return result, fmt.Errorf("invalid shipping type")
	}
	shippingType := models.TradeShippingType(shippingTypeValue)
	if !shippingType.Valid() {
		return result, fmt.Errorf("invalid shipping type")
	}
	if shippingType.NeedsLocation() && utf8.RuneCountInString(location) < 2 {
		return result, fmt.Errorf("location is required for meetup")
	}

	result = models.TradeCommonItem{
		Brand:            brand,
		Price:            price,
		PriceType:        priceType,
		Currency:         currency,
		ProductCondition: productCondition,
		Location:         location,
		ShippingType:     shippingType,
		Status:           models.TRADE_AVAILABLE,
	}
	return result, nil
}

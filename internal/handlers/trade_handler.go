package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
)

type TradeHandler interface {
	AddFavoriteHandler(c fiber.Ctx) error
	RatingSellerHandler(c fiber.Ctx) error
	TradeListHandler(c fiber.Ctx) error
	TradeModifyHandler(c fiber.Ctx) error
	TradeViewHandler(c fiber.Ctx) error
	TradeWriteHandler(c fiber.Ctx) error
	UpdateStatusHandler(c fiber.Ctx) error
}

type TsboardTradeHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardTradeHandler(service *services.Service) *TsboardTradeHandler {
	return &TsboardTradeHandler{service: service}
}

// 찜하기 등록하기 핸들러
func (h *TsboardTradeHandler) AddFavoriteHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

// 판매 완료 후 구매자의 별점과 거래 후기 등록 핸들러
func (h *TsboardTradeHandler) RatingSellerHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

// 거래 목록 가져오기 핸들러
func (h *TsboardTradeHandler) TradeListHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

// 거래 내용 수정하기 핸들러
func (h *TsboardTradeHandler) TradeModifyHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

// 거래 보기 핸들러
func (h *TsboardTradeHandler) TradeViewHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

// 새 거래 작성하기 핸들러
func (h *TsboardTradeHandler) TradeWriteHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

// 거래 상태 변경 핸들러
func (h *TsboardTradeHandler) UpdateStatusHandler(c fiber.Ctx) error {
	// TODO
	return fmt.Errorf("not implemented yet")
}

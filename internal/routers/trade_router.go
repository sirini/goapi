package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 물품 거래 게시판과 상호작용에 필요한 라우터들 등록
func RegisterTradeRouters(api fiber.Router, h *handlers.Handler) {
	trade := api.Group("/trade")
	trade.Get("/list", h.Trade.TradeListHandler)
	trade.Get("/view", h.Trade.TradeViewHandler)

	protected := trade.Group("/", middlewares.JWTMiddleware())
	protected.Post("/modify", h.Trade.TradeModifyHandler)
	protected.Post("/write", h.Trade.TradeWriteHandler)
	protected.Patch("/update/status", h.Trade.UpdateStatusHandler)
}

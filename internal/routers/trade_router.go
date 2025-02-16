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

	trade.Patch("/favorite", h.Trade.AddFavoriteHandler, middlewares.JWTMiddleware())
	trade.Post("/rating/seller", h.Trade.RatingSellerHandler, middlewares.JWTMiddleware())
	trade.Post("/modify", h.Trade.TradeModifyHandler, middlewares.JWTMiddleware())
	trade.Post("/write", h.Trade.TradeWriteHandler, middlewares.JWTMiddleware())
	trade.Patch("/update/status", h.Trade.UpdateStatusHandler, middlewares.JWTMiddleware())
}

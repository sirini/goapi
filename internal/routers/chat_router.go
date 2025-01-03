package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 쪽지 관련 라우터들 등록
func RegisterChatRouters(api fiber.Router, h *handlers.Handler) {
	chat := api.Group("/chat")
	chat.Get("/list", h.Chat.LoadChatListHandler, middlewares.JWTMiddleware())
	chat.Get("/history", h.Chat.LoadChatHistoryHandler, middlewares.JWTMiddleware())
	chat.Post("/save", h.Chat.SaveChatHandler, middlewares.JWTMiddleware())
}

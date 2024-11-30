package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 알림 관련 라우터들 등록
func RegisterNotiRouters(api fiber.Router, h *handlers.Handler) {
	noti := api.Group("/noti")
	noti.Get("/load", h.Noti.LoadNotiListHandler, middlewares.JWTMiddleware())
	noti.Patch("/checked", h.Noti.CheckedAllNotiHandler, middlewares.JWTMiddleware())
}

package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 사용자 관련 라우터들 등록
func RegisterUserRouters(api fiber.Router, h *handlers.Handler) {
	user := api.Group("/user")
	user.Get("/load/user/info", h.User.LoadUserInfoHandler)

	user.Post("/report", h.User.ReportUserHandler, middlewares.JWTMiddleware())
	user.Get("/load/permission", h.User.LoadUserPermissionHandler, middlewares.JWTMiddleware())
	user.Post("/manage/user", h.User.ManageUserPermissionHandler, middlewares.JWTMiddleware())
}

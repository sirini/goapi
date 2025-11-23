package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 홈화면 및 SEO용 라우터들 등록
func RegisterHomeRouters(api fiber.Router, h *handlers.Handler) {
	home := api.Group("/home")
	home.Get("/tsboard", h.Home.ShowVersionHandler)
	home.Get("/visit", h.Home.CountingVisitorHandler)
	home.Get("/latest", h.Home.LoadAllPostsHandler)
	home.Get("/latest/post", h.Home.LoadPostsByIdHandler)
	home.Get("/sidebar/links", h.Home.LoadSidebarLinkHandler)

	// 알림용 라우터들
	noti := home.Group("/noti")
	noti.Get("/load", h.Noti.LoadNotiListHandler, middlewares.JWTMiddleware())
	noti.Patch("/checked", h.Noti.CheckedAllNotiHandler, middlewares.JWTMiddleware())
	noti.Patch("/checked/:notiUid", h.Noti.CheckedSingleNotiHandler, middlewares.JWTMiddleware())
}

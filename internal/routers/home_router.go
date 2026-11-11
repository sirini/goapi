package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/securitylog"
	"time"
)

// 홈화면 및 SEO용 라우터들 등록
func RegisterHomeRouters(api fiber.Router, h *handlers.Handler) {
	home := api.Group("/home")
	home.Post("/page-visit", limiter.New(limiter.Config{Max: 120, Expiration: time.Minute, KeyGenerator: securitylog.RequestPeer}), securitylog.VisitHandler(h.CanAuthenticate))
	home.Get("/nubo", h.Home.ShowVersionHandler)
	home.Get("/visit", h.Home.CountingVisitorHandler)
	home.Get("/latest", h.Home.LoadAllPostsHandler)
	home.Get("/latest/:id", h.Home.LoadPostsByIdHandler)
	home.Get("/sidebar/links", h.Home.LoadSidebarLinkHandler)

	// 알림용 라우터들
	noti := home.Group("/noti", middlewares.JWTMiddleware(h.CanAuthenticate))
	noti.Get("/load", h.Noti.LoadNotiListHandler)
	noti.Patch("/checked", h.Noti.CheckedAllNotiHandler)
	noti.Patch("/checked/:notiUid", h.Noti.CheckedSingleNotiHandler)
}

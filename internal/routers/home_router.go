package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
)

// 홈화면 및 SEO용 라우터들 등록
func RegisterHomeRouters(api fiber.Router, h *handlers.Handler) {
	home := api.Group("/home")
	home.Get("/tsboard", h.Home.ShowVersionHandler)
	home.Get("/visit", h.Home.CountingVisitorHandler)
	home.Get("/latest", h.Home.LoadAllPostsHandler)
	home.Get("/latest/post", h.Home.LoadPostsByIdHandler)
	home.Get("/sidebar/links", h.Home.LoadSidebarLinkHandler)

	seo := api.Group("/seo")
	seo.Get("/main.html", h.Home.LoadMainPageHandler)
	seo.Get("/sitemap.xml", h.Home.LoadSitemapHandler)
}

package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
)

// 게시판과 상호작용에 필요한 라우터들 등록
func RegisterBlogRouters(api fiber.Router, h *handlers.Handler) {
		rss := api.Group("/rss")
	rss.Get("/:id", h.Board.BoardListHandler)
}
package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 게시판과 상호작용에 필요한 라우터들 등록
func RegisterBoardRouters(api fiber.Router, h *handlers.Handler) {
	board := api.Group("/board")
	board.Get("/list", h.Board.BoardListHandler)
	board.Get("/view", h.Board.BoardViewHandler)
	board.Get("/tag/recent", h.Board.BoardRecentTagListHandler)
	board.Get("/user/latest", h.Board.LatestUserContentHandler)
	board.Get("/transfer", h.Board.TransferHandler)

	protected := board.Group("/", middlewares.JWTMiddleware())
	protected.Get("/download", h.Board.DownloadHandler)
	protected.Get("/move/list", h.Board.ListForMoveHandler)
	protected.Patch("/like", h.Board.LikePostHandler)
	protected.Post("/move/apply", h.Board.MovePostHandler)
	protected.Delete("/remove/post", h.Board.RemovePostHandler)
}

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
	board.Get("/photo/list", h.Board.GalleryListHandler)
	board.Get("/photo/view", h.Board.GalleryLoadPhotoHandler)

	board.Get("/download", h.Board.DownloadHandler, middlewares.JWTMiddleware())
	board.Get("/config", h.Board.GetEditorConfigHandler, middlewares.JWTMiddleware())
	board.Get("/move/list", h.Board.ListForMoveHandler, middlewares.JWTMiddleware())
	board.Patch("/like", h.Board.LikePostHandler, middlewares.JWTMiddleware())
	board.Get("/load/images", h.Board.LoadInsertImageHandler, middlewares.JWTMiddleware())
	board.Get("/load/post", h.Board.LoadPostHandler, middlewares.JWTMiddleware())
	board.Put("/move/apply", h.Board.MovePostHandler, middlewares.JWTMiddleware())
	board.Patch("/modify", h.Board.ModifyPostHandler, middlewares.JWTMiddleware())
	board.Delete("/remove/attached", h.Board.RemoveAttachedFileHandler, middlewares.JWTMiddleware())
	board.Delete("/remove/post", h.Board.RemovePostHandler, middlewares.JWTMiddleware())
	board.Delete("/remove/image", h.Board.RemoveInsertImageHandler, middlewares.JWTMiddleware())
	board.Get("/tag/suggestion", h.Board.SuggestionHashtagHandler, middlewares.JWTMiddleware())
	board.Post("/upload/images", h.Board.UploadInsertImageHandler, middlewares.JWTMiddleware())
	board.Post("/write", h.Board.WritePostHandler, middlewares.JWTMiddleware())
}

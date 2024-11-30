package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 댓글 관련 라우터들 등록하기
func RegisterCommentRouters(api fiber.Router, h *handlers.Handler) {
	comment := api.Group("/comment")
	comment.Get("/list", h.Comment.CommentListHandler)

	comment.Patch("/like", h.Comment.LikeCommentHandler, middlewares.JWTMiddleware())
	comment.Patch("/modify", h.Comment.ModifyCommentHandler, middlewares.JWTMiddleware())
	comment.Delete("/remove", h.Comment.RemoveCommentHandler, middlewares.JWTMiddleware())
	comment.Post("/reply", h.Comment.ReplyCommentHandler, middlewares.JWTMiddleware())
	comment.Post("/write", h.Comment.WriteCommentHandler, middlewares.JWTMiddleware())
}

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

	protected := comment.Group("/", middlewares.JWTMiddleware())
	protected.Patch("/like", h.Comment.LikeCommentHandler)
	protected.Patch("/modify", h.Comment.ModifyCommentHandler)
	protected.Delete("/remove", h.Comment.RemoveCommentHandler)
	protected.Post("/reply", h.Comment.ReplyCommentHandler)
	protected.Post("/write", h.Comment.WriteCommentHandler)
}

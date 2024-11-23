package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 일반 댓글 라우터들 등록하기
func SetupCommentRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/board/comment", handlers.CommentListHandler(s))
}

// 댓글 관련 중에서 로그인이 필요한 라우터들 등록하기
func SetupLoggedInCommentRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("PATCH /goapi/board/like/comment", middlewares.AuthMiddleware(handlers.LikeCommentHandler(s)))
}

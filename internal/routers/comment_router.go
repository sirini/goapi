package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// 일반 댓글 라우터들 등록하기
func SetupCommentRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/board/comment", handlers.CommentListHandler(s))
}

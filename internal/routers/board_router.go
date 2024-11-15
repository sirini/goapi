package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 일반 게시판 라우터들 등록하기
func SetupBoardRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/board/list", handlers.LoadBoardListHandler(s))
	mux.HandleFunc("GET /goapi/board/view", handlers.LoadBoardViewHandler(s))
}

// 로그인이 필요한 게시판 라우터들 등록하기
func SetupLoggedInBoardRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("PATCH /goapi/board/like/post", middlewares.AuthMiddleware(handlers.LikePostHandler(s)))
}

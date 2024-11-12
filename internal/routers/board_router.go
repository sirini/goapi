package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// 게시판 라우터들 등록하기
func SetupBoardRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/board/list", handlers.LoadBoardListHandler(s))
	mux.HandleFunc("GET /goapi/board/view", handlers.LoadBoardViewHandler(s))
}

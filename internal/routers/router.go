package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/services"
)

// 모든 라우터들을 등록하기
func SetupRoutes(mux *http.ServeMux, s *services.Service) {
	SetupHomeRouter(mux, s)
	SetupUserRouter(mux, s)
	SetupAuthRouter(mux, s)
	SetupOAuthRouter(mux, s)
	SetupBoardRouter(mux, s)

	SetupLoggedInUserRouter(mux, s)
	SetupLoggedInAuthRouter(mux, s)
	SetupLoggedInChatRouter(mux, s)
	SetupLoggedInNotiRouter(mux, s)
	SetupLoggedInBoardRouter(mux, s)
}

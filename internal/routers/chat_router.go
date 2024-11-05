package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 쪽지 남기기 관련 라우터 셋업
func SetupLoggedInChatRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("GET /goapi/user/load/chat/list", middlewares.AuthMiddleware(handlers.LoadChatListHandler(s)))
	mux.Handle("GET /goapi/user/load/chat/history", middlewares.AuthMiddleware(handlers.LoadChatHistoryHandler(s)))
	mux.Handle("POST /goapi/user/save/chat", middlewares.AuthMiddleware(handlers.SaveChatHandler(s)))
}

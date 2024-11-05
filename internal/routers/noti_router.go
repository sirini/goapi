package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 알림 확인용 라우터 셋업
func SetupLoggedInNotiRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("GET /goapi/home/load/notification", middlewares.AuthMiddleware(handlers.LoadNotiListHandler(s)))
	mux.Handle("PATCH /goapi/home/checked/notification", middlewares.AuthMiddleware(handlers.CheckedAllNoti(s)))
}

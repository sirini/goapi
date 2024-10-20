package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// 모든 라우터들을 등록하기
func SetupRoutes(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/tsboard", handlers.Version)
	mux.HandleFunc("GET /goapi/user/load/user/info", handlers.LoadUserInfoHandler(s))
	mux.HandleFunc("POST /goapi/user/report", handlers.ReportUserHandler(s))
}

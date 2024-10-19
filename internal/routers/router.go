package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// 모든 라우터들을 등록하기
func SetupRoutes(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/tsboard", handlers.Version)
	mux.HandleFunc("GET /goapi/load/user/info", handlers.LoadUserInfoHandler(s))
}

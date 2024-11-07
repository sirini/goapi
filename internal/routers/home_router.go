package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// TSBOARD : GOAPI 버전 확인 라우터 셋업
func SetupVersionRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/home/tsboard", handlers.ShowVersionHandler)
	mux.HandleFunc("GET /goapi/home/visit", handlers.CountingVisitorHandler(s))
	mux.HandleFunc("GET /goapi/home/sidebar/links", handlers.LoadSidebarLinkHandler(s))
}

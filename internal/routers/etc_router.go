package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// TSBOARD : GOAPI 버전 확인 라우터 셋업
func SetupVersionRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/tsboard", handlers.Version)
}

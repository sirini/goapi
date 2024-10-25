package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/services"
)

// 모든 라우터들을 등록하기
func SetupRoutes(mux *http.ServeMux, s *services.Service) {
	SetupVersionRouter(mux, s)
	SetupUserRouter(mux, s)
	SetupAuthUserRouter(mux, s)
}

package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 사용자 관련 라우터 중 비 로그인 라우터 셋업
func SetupUserRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/user/load/user/info", handlers.LoadUserInfoHandler(s))
	mux.HandleFunc("POST /goapi/user/checkemail", handlers.CheckEmailHandler(s))
	mux.HandleFunc("POST /goapi/user/checkname", handlers.CheckNameHandler(s))
	mux.HandleFunc("POST /goapi/user/verify", handlers.VerifyCodeHandler(s))
}

// 사용자 관련 라우터 중 로그인 필요한 라우터 셋업
func SetupLoggedInUserRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("POST /goapi/user/report", middlewares.AuthMiddleware(handlers.ReportUserHandler(s)))
	mux.Handle("GET /goapi/user/load/permission", middlewares.AuthMiddleware(handlers.LoadUserPermissionHandler(s)))
}

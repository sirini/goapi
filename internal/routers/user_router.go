package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 사용자 관련 라우터 중 인증 불필요한 라우터 셋업
func SetupUserRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/user/load/user/info", handlers.LoadUserInfoHandler(s))
	mux.HandleFunc("POST /goapi/user/signin", handlers.SigninHandler(s))
	mux.HandleFunc("POST /goapi/user/signup", handlers.SignupHandler(s))
	mux.HandleFunc("POST /goapi/user/checkemail", handlers.CheckEmailHandler(s))
	mux.HandleFunc("POST /goapi/user/checkname", handlers.CheckNameHandler(s))
	mux.HandleFunc("POST /goapi/user/verify", handlers.VerifyCodeHandler(s))
	mux.HandleFunc("POST /goapi/user/reset/password", handlers.ResetPasswordHandler(s))
}

// 사용자 관련 라우터 중 로그인이 요구되는 라우터 셋업
func SetupAuthUserRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("POST /goapi/user/report", middlewares.AuthMiddleware(handlers.ReportUserHandler(s)))
}

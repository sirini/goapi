package routers

import (
	"fmt"
	"net/http"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// 사용자 인증 관련 비 로그인 라우터 셋업
func SetupAuthRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("POST /goapi/auth/signin", handlers.SigninHandler(s))
	mux.HandleFunc("POST /goapi/auth/signup", handlers.SignupHandler(s))
	mux.HandleFunc("POST /goapi/auth/reset/password", handlers.ResetPasswordHandler(s))
}

// 사용자 정보 가져오기 등 로그인 필요한 라우터 셋업
func SetupLoggedInAuthRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("GET /goapi/auth/load", middlewares.AuthMiddleware(handlers.LoadMyInfoHandler(s)))
	mux.Handle("PATCH /goapi/auth/update", middlewares.AuthMiddleware(handlers.UpdateMyInfoHandler(s)))
	mux.Handle("POST /goapi/auth/logout", middlewares.AuthMiddleware(handlers.LogoutHandler(s)))
}

// OAuth 사용자 로그인 관련 라우터 셋업
func SetupOAuthRouter(mux *http.ServeMux, s *services.Service) {
	cfgGoogle := oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/goapi/auth/google/callback", configs.Env.URL),
		ClientID:     configs.Env.OAuthGoogleID,
		ClientSecret: configs.Env.OAuthGoogleSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	cfgNaver := oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/goapi/auth/naver/callback", configs.Env.URL),
		ClientID:     configs.Env.OAuthNaverID,
		ClientSecret: configs.Env.OAuthNaverSecret,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://nid.naver.com/oauth2.0/authorize",
			TokenURL: "https://nid.naver.com/oauth2.0/token",
		},
	}
	cfgKakao := oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/goapi/auth/kakao/callback", configs.Env.URL),
		ClientID:     configs.Env.OAuthKakaoID,
		ClientSecret: configs.Env.OAuthKakaoSecret,
		Scopes:       []string{"account_email", "profile_image", "profile_nickname"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://kauth.kakao.com/oauth/authorize",
			TokenURL: "https://kauth.kakao.com/oauth/token",
		},
	}

	mux.HandleFunc("GET /goapi/auth/google/request", handlers.GoogleOAuthRequestHandler(s, cfgGoogle))
	mux.HandleFunc("GET /goapi/auth/google/callback", handlers.GoogleOAuthCallbackHandler(s, cfgGoogle))
	mux.HandleFunc("GET /goapi/auth/naver/request", handlers.NaverOAuthRequestHandler(s))
	mux.HandleFunc("GET /goapi/auth/naver/callback", handlers.NaverOAuthCallbackHandler(s, cfgNaver))
	mux.HandleFunc("GET /goapi/auth/kakao/request", handlers.KakaoOAuthRequestHandler(s, cfgKakao))
	mux.HandleFunc("GET /goapi/auth/kakao/callback", handlers.KakaoOAuthCallbackHandler(s, cfgKakao))
	mux.HandleFunc("GET /goapi/auth/oauth/userinfo", handlers.RequestUserInfoHandler(s))
}

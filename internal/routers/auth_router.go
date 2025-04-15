package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 사용자 인증 관련 라우터들 등록
func RegisterAuthRouters(api fiber.Router, h *handlers.Handler) {
	auth := api.Group("/auth")
	auth.Post("/signin", h.Auth.SigninHandler)
	auth.Post("/signup", h.Auth.SignupHandler)
	auth.Post("/reset/password", h.Auth.ResetPasswordHandler)
	auth.Post("/refresh", h.Auth.RefreshAccessTokenHandler)
	auth.Post("/checkemail", h.Auth.CheckEmailHandler)
	auth.Post("/checkname", h.Auth.CheckNameHandler)
	auth.Post("/verify", h.Auth.VerifyCodeHandler)

	auth.Get("/load", h.Auth.LoadMyInfoHandler, middlewares.JWTMiddleware())
	auth.Post("/logout", h.Auth.LogoutHandler, middlewares.JWTMiddleware())
	auth.Patch("/update", h.Auth.UpdateMyInfoHandler, middlewares.JWTMiddleware())

	// OAuth용 라우터들
	auth.Get("/google/request", h.OAuth2.GoogleOAuthRequestHandler)
	auth.Get("/google/callback", h.OAuth2.GoogleOAuthCallbackHandler)
	auth.Get("/naver/request", h.OAuth2.NaverOAuthRequestHandler)
	auth.Get("/naver/callback", h.OAuth2.NaverOAuthCallbackHandler)
	auth.Get("/kakao/request", h.OAuth2.KakaoOAuthRequestHandler)
	auth.Get("/kakao/callback", h.OAuth2.KakaoOAuthCallbackHandler)
	auth.Get("/oauth/userinfo", h.OAuth2.RequestUserInfoHandler)

	// Android OAuth용 라우터
	auth.Post("/android/google", h.OAuth2.AndroidGoogleOAuthHandler)
}

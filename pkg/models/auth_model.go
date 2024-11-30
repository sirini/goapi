package models

import (
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// 회원가입 시 리턴 타입
type SignupResult struct {
	Sendmail bool `json:"sendmail"`
	Target   uint `json:"target"`
}

// 인증 완료하기 파라미터
type VerifyParameter struct {
	Target   uint
	Code     string
	Id       string
	Password string
	Name     string
}

// 비밀번호 초기화 시 리턴 타입
type ResetPasswordResult struct {
	Sendmail bool `json:"sendmail"`
}

// 구글 OAuth 응답
type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// 네이버 OAuth 응답
type NaverUser struct {
	Response struct {
		Email        string `json:"email"`
		Nickname     string `json:"nickname"`
		ProfileImage string `json:"profile_image"`
	} `json:"response"`
}

// 카카오 OAuth 응답
type KakaoUser struct {
	ID           int64 `json:"id"`
	KakaoAccount struct {
		Email   string `json:"email"`
		Profile struct {
			Nickname        string `json:"nickname"`
			ProfileImageUrl string `json:"profile_image_url"`
		} `json:"profile"`
	} `json:"kakao_account"`
}

// 인증 메일 발송에 필요한 파라미터 정의
type SignupParameter struct {
	ID       string
	Password string
	Name     string
	Hostname string
}

// JWT 컨텍스트 키값 설정
type ContextKey string

var JwtClaimsKey = ContextKey("jwtClaims")

// Google OAuth2 설정값 정의
var GoogleOAuth2Config = oauth2.Config{
	RedirectURL:  fmt.Sprintf("%s/goapi/auth/google/callback", configs.Env.URL),
	ClientID:     configs.Env.OAuthGoogleID,
	ClientSecret: configs.Env.OAuthGoogleSecret,
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// Naver OAuth2 설정값 정의
var NaverOAuth2Config = oauth2.Config{
	RedirectURL:  fmt.Sprintf("%s/goapi/auth/naver/callback", configs.Env.URL),
	ClientID:     configs.Env.OAuthNaverID,
	ClientSecret: configs.Env.OAuthNaverSecret,
	Scopes:       []string{},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://nid.naver.com/oauth2.0/authorize",
		TokenURL: "https://nid.naver.com/oauth2.0/token",
	},
}

// Kakao OAuth2 설정값 정의
var KakaoOAuth2Config = oauth2.Config{
	RedirectURL:  fmt.Sprintf("%s/goapi/auth/kakao/callback", configs.Env.URL),
	ClientID:     configs.Env.OAuthKakaoID,
	ClientSecret: configs.Env.OAuthKakaoSecret,
	Scopes:       []string{"account_email", "profile_image", "profile_nickname"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://kauth.kakao.com/oauth/authorize",
		TokenURL: "https://kauth.kakao.com/oauth/token",
	},
}

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuth2Handler interface {
	AndroidGoogleOAuthHandler(c fiber.Ctx) error
	GoogleOAuthRequestHandler(c fiber.Ctx) error
	GoogleOAuthCallbackHandler(c fiber.Ctx) error
	NaverOAuthRequestHandler(c fiber.Ctx) error
	NaverOAuthCallbackHandler(c fiber.Ctx) error
	KakaoOAuthRequestHandler(c fiber.Ctx) error
	KakaoOAuthCallbackHandler(c fiber.Ctx) error
	UtilRegisterUser(id string, name string, profile string) uint
	UtilFinishLogin(c fiber.Ctx, userUid uint) error
}

type NuboOAuth2Handler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboOAuth2Handler(service *services.Service) *NuboOAuth2Handler {
	return &NuboOAuth2Handler{service: service}
}

// 구글 안드로이드 앱 OAuth 콜백 핸들러
func (h *NuboOAuth2Handler) AndroidGoogleOAuthHandler(c fiber.Ctx) error {
	if configs.Env.OAuthGoogleID == "" {
		return utils.Err(c, "google oauth is not configured", models.CODE_FAILED_OPERATION)
	}
	idToken := c.FormValue("id_token")
	if len(idToken) < 1 {
		return utils.Err(c, "id_token is empty", models.CODE_INVALID_PARAMETER)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken))
	if err != nil || resp.StatusCode != http.StatusOK {
		return utils.Err(c, "invalid google token", models.CODE_INVALID_TOKEN)
	}
	defer resp.Body.Close()

	var userInfo models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	if !validGoogleIDTokenInfo(userInfo, configs.Env.OAuthGoogleID) {
		return utils.Err(c, "invalid google token claims", models.CODE_INVALID_TOKEN)
	}

	userUid := h.UtilRegisterUser(userInfo.Email, userInfo.Name, userInfo.Picture)
	if userUid < 1 {
		return utils.Err(c, "failed to registrate a user", models.CODE_FAILED_OPERATION)
	}
	if !h.service.Auth.CanAuthenticate(userUid) {
		return utils.Err(c, "this account is not allowed to sign in", models.CODE_NO_PERMISSION)
	}

	auth, refresh := h.service.OAuth.GenerateTokens(userUid)
	h.service.OAuth.SaveRefreshToken(userUid, refresh)

	user := h.service.OAuth.GetUserInfo(userUid)
	user.Token = auth
	user.Refresh = refresh
	return utils.Ok(c, user)
}

// 구글 OAuth 로그인을 위해 리다이렉트
func (h *NuboOAuth2Handler) GoogleOAuthRequestHandler(c fiber.Ctx) error {
	state := uuid.NewString()
	utils.SaveCookie(c, models.OAUTH_STATE, state, 1)
	googleConfig := googleOAuthConfig()
	url := googleConfig.AuthCodeURL(state)
	return c.Redirect().To(url)
}

// 구글 OAuth 콜백 핸들러
func (h *NuboOAuth2Handler) GoogleOAuthCallbackHandler(c fiber.Ctx) error {
	if configs.Env.OAuthGoogleID == "" {
		return c.Redirect().To(configs.Env.Domain)
	}

	googleConfig := googleOAuthConfig()
	token, err := utils.OAuth2ExchangeToken(c, googleConfig)
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}

	client := googleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}
	defer resp.Body.Close()

	var userInfo models.GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}

	userUid := h.UtilRegisterUser(userInfo.Email, userInfo.Name, userInfo.Picture)
	if userUid < 1 {
		return c.Redirect().To(configs.Env.Domain)
	}

	return h.UtilFinishLogin(c, userUid)
}

// 네이버 OAuth 로그인을 위해 리다이렉트
func (h *NuboOAuth2Handler) NaverOAuthRequestHandler(c fiber.Ctx) error {
	state := uuid.NewString()
	utils.SaveCookie(c, models.OAUTH_STATE, state, 1)
	naverRedirectURL := oauthRedirectURL("naver")
	url := fmt.Sprintf(
		"https://nid.naver.com/oauth2.0/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		configs.Env.OAuthNaverID,
		naverRedirectURL,
		state,
	)
	return c.Redirect().To(url)
}

// 네이버 OAuth 콜백 핸들러
func (h *NuboOAuth2Handler) NaverOAuthCallbackHandler(c fiber.Ctx) error {
	if configs.Env.OAuthNaverID == "" {
		return c.Redirect().To(configs.Env.Domain)
	}

	code := c.FormValue("code")
	state := c.FormValue("state")
	naverRedirectURL := oauthRedirectURL("naver")

	cookie := c.Cookies(models.OAUTH_STATE)
	if !utils.OAuthStateMatches(cookie, state) {
		return c.Redirect().To(configs.Env.Domain)
	}
	c.ClearCookie(models.OAUTH_STATE)

	apiURL := fmt.Sprintf(
		"https://nid.naver.com/oauth2.0/token?grant_type=authorization_code&client_id=%s&client_secret=%s&redirect_uri=%s&code=%s&state=%s",
		configs.Env.OAuthNaverID,
		configs.Env.OAuthNaverSecret,
		url.QueryEscape(naverRedirectURL),
		code,
		state,
	)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}
	req.Header.Set("X-Naver-Client-Id", configs.Env.OAuthNaverID)
	req.Header.Set("X-Naver-Client-Secret", configs.Env.OAuthNaverSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.Redirect().To(configs.Env.Domain)
	}
	defer resp.Body.Close()

	var tokenResp map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok || accessToken == "" {
		return c.Redirect().To(configs.Env.Domain)
	}

	naverConfig := oauth2.Config{
		RedirectURL:  naverRedirectURL,
		ClientID:     configs.Env.OAuthNaverID,
		ClientSecret: configs.Env.OAuthNaverSecret,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://nid.naver.com/oauth2.0/authorize",
			TokenURL: "https://nid.naver.com/oauth2.0/token",
		},
	}

	client = naverConfig.Client(context.Background(), &oauth2.Token{
		AccessToken: accessToken,
	})

	resp, err = client.Get("https://openapi.naver.com/v1/nid/me")
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.Redirect().To(configs.Env.Domain)
	}
	defer resp.Body.Close()

	var userInfo models.NaverUser
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}

	id := userInfo.Response.Email
	name := userInfo.Response.Nickname
	profile := userInfo.Response.ProfileImage
	userUid := h.UtilRegisterUser(id, name, profile)
	if userUid < 1 {
		return c.Redirect().To(configs.Env.Domain)
	}

	return h.UtilFinishLogin(c, userUid)
}

// 카카오 OAuth 로그인을 위해 리다이렉트
func (h *NuboOAuth2Handler) KakaoOAuthRequestHandler(c fiber.Ctx) error {
	state := uuid.NewString()
	utils.SaveCookie(c, models.OAUTH_STATE, state, 1)
	kakaoConfig := kakaoOAuthConfig()
	url := kakaoConfig.AuthCodeURL(state)
	return c.Redirect().To(url)
}

// 카카오 OAuth 콜백 핸들러
func (h *NuboOAuth2Handler) KakaoOAuthCallbackHandler(c fiber.Ctx) error {
	if configs.Env.OAuthKakaoID == "" {
		return c.Redirect().To(configs.Env.Domain)
	}

	kakaoConfig := kakaoOAuthConfig()
	token, err := utils.OAuth2ExchangeToken(c, kakaoConfig)
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}

	client := kakaoConfig.Client(context.Background(), token)
	resp, err := client.Get("https://kapi.kakao.com/v2/user/me")
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}
	defer resp.Body.Close()

	var userInfo models.KakaoUser
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return c.Redirect().To(configs.Env.Domain)
	}

	id := userInfo.KakaoAccount.Email
	name := userInfo.KakaoAccount.Profile.Nickname
	profile := userInfo.KakaoAccount.Profile.ProfileImageUrl
	userUid := h.UtilRegisterUser(id, name, profile)
	if userUid < 1 {
		return c.Redirect().To(configs.Env.Domain)
	}
	return h.UtilFinishLogin(c, userUid)
}

// 이미 등록된 사용자인지 확인하고 필요 시 등록 후 고유번호 반환
func (h *NuboOAuth2Handler) UtilRegisterUser(id string, name string, profile string) uint {
	isRegistered := h.service.Auth.CheckEmailExists(id)
	var userUid uint
	if !isRegistered {
		userUid = h.service.OAuth.RegisterOAuthUser(id, name, profile)
	} else {
		userUid = h.service.OAuth.GetUserUid(id)
	}
	return userUid
}

// 토큰 저장 및 쿠키에 사용자 정보 전달
func (h *NuboOAuth2Handler) UtilFinishLogin(c fiber.Ctx, userUid uint) error {
	if !h.service.Auth.CanAuthenticate(userUid) {
		return c.Redirect().To(configs.Env.Domain)
	}
	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	auth, refresh := h.service.OAuth.GenerateTokens(userUid)
	h.service.OAuth.SaveRefreshToken(userUid, refresh)

	utils.SaveCookie(c, models.AUTH_TOKEN, auth, accessHours)
	utils.SaveCookie(c, models.REFRESH_TOKEN, refresh, refreshDays*24)

	return c.Redirect().To(configs.Env.Domain)
}

func oauthRedirectURL(provider string) string {
	return fmt.Sprintf("%s/%s/auth/%s/callback", configs.Env.Domain, configs.Env.GoapiBase, provider)
}

func googleOAuthConfig() oauth2.Config {
	return oauth2.Config{
		RedirectURL:  oauthRedirectURL("google"),
		ClientID:     configs.Env.OAuthGoogleID,
		ClientSecret: configs.Env.OAuthGoogleSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func kakaoOAuthConfig() oauth2.Config {
	return oauth2.Config{
		RedirectURL:  oauthRedirectURL("kakao"),
		ClientID:     configs.Env.OAuthKakaoID,
		ClientSecret: configs.Env.OAuthKakaoSecret,
		Scopes:       []string{"account_email", "profile_image", "profile_nickname"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://kauth.kakao.com/oauth/authorize",
			TokenURL: "https://kauth.kakao.com/oauth/token",
		},
	}
}

func validGoogleIDTokenInfo(user models.GoogleUser, clientID string) bool {
	return clientID != "" && user.Audience == clientID && user.Email != "" && user.EmailVerified == "true"
}

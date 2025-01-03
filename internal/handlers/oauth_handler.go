package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
	GoogleOAuthRequestHandler(c fiber.Ctx) error
	GoogleOAuthCallbackHandler(c fiber.Ctx) error
	NaverOAuthRequestHandler(c fiber.Ctx) error
	NaverOAuthCallbackHandler(c fiber.Ctx) error
	KakaoOAuthRequestHandler(c fiber.Ctx) error
	KakaoOAuthCallbackHandler(c fiber.Ctx) error
	RequestUserInfoHandler(c fiber.Ctx) error
	UtilRegisterUser(id string, name string, profile string) uint
	UtilFinishLogin(c fiber.Ctx, userUid uint) error
}

type TsboardOAuth2Handler struct {
	service          *services.Service
	googleConfig     oauth2.Config
	naverRedirectURL string
	naverConfig      oauth2.Config
	kakaoConfig      oauth2.Config
}

// services.Service 주입 받기
func NewTsboardOAuth2Handler(service *services.Service) *TsboardOAuth2Handler {
	return &TsboardOAuth2Handler{service: service}
}

// 구글 OAuth 로그인을 위해 리다이렉트
func (h *TsboardOAuth2Handler) GoogleOAuthRequestHandler(c fiber.Ctx) error {
	state := uuid.New().String()[:10]
	utils.SaveCookie(c, "tsboard_oauth_state", state, 1)

	h.googleConfig = oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s%s/goapi/auth/google/callback", configs.Env.URL, configs.Env.URLPrefix),
		ClientID:     configs.Env.OAuthGoogleID,
		ClientSecret: configs.Env.OAuthGoogleSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	url := h.googleConfig.AuthCodeURL(state)
	return c.Redirect().To(url)
}

// 구글 OAuth 콜백 핸들러
func (h *TsboardOAuth2Handler) GoogleOAuthCallbackHandler(c fiber.Ctx) error {
	redirectPath := fmt.Sprintf("%s%s", configs.Env.URL, configs.Env.URLPrefix)
	if configs.Env.OAuthGoogleID == "" {
		return c.Redirect().To(redirectPath)
	}

	token, err := utils.OAuth2ExchangeToken(c, h.googleConfig)
	if err != nil {
		return c.Redirect().To(redirectPath)
	}

	client := h.googleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Redirect().To(redirectPath)
	}
	defer resp.Body.Close()

	var userInfo models.GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return c.Redirect().To(redirectPath)
	}

	id := userInfo.Email
	name := userInfo.Name
	profile := userInfo.Picture
	userUid := h.UtilRegisterUser(id, name, profile)
	if userUid < 1 {
		return c.Redirect().To(redirectPath)
	}

	return h.UtilFinishLogin(c, userUid)
}

// 네이버 OAuth 로그인을 위해 리다이렉트
func (h *TsboardOAuth2Handler) NaverOAuthRequestHandler(c fiber.Ctx) error {
	state := uuid.New().String()[:10]
	utils.SaveCookie(c, "tsboard_oauth_state", state, 1)

	h.naverRedirectURL = fmt.Sprintf("%s%s/goapi/auth/naver/callback", configs.Env.URL, configs.Env.URLPrefix)
	url := fmt.Sprintf(
		"https://nid.naver.com/oauth2.0/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		configs.Env.OAuthNaverID,
		h.naverRedirectURL,
		state,
	)
	return c.Redirect().To(url)
}

// 네이버 OAuth 콜백 핸들러
func (h *TsboardOAuth2Handler) NaverOAuthCallbackHandler(c fiber.Ctx) error {
	redirectPath := fmt.Sprintf("%s%s", configs.Env.URL, configs.Env.URLPrefix)
	if configs.Env.OAuthNaverID == "" {
		return c.Redirect().To(redirectPath)
	}

	code := c.FormValue("code")
	state := c.FormValue("state")

	cookie := c.Cookies("tsboard_oauth_state")
	if cookie != state {
		return c.Redirect().To(redirectPath)
	}

	apiURL := fmt.Sprintf(
		"https://nid.naver.com/oauth2.0/token?grant_type=authorization_code&client_id=%s&client_secret=%s&redirect_uri=%s&code=%s&state=%s",
		configs.Env.OAuthNaverID,
		configs.Env.OAuthNaverSecret,
		url.QueryEscape(h.naverRedirectURL),
		code,
		state,
	)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return c.Redirect().To(redirectPath)
	}
	req.Header.Set("X-Naver-Client-Id", configs.Env.OAuthNaverID)
	req.Header.Set("X-Naver-Client-Secret", configs.Env.OAuthNaverSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.Redirect().To(redirectPath)
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return c.Redirect().To(redirectPath)
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok || accessToken == "" {
		return c.Redirect().To(redirectPath)
	}

	h.naverConfig = oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/goapi/auth/naver/callback", redirectPath),
		ClientID:     configs.Env.OAuthNaverID,
		ClientSecret: configs.Env.OAuthNaverSecret,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://nid.naver.com/oauth2.0/authorize",
			TokenURL: "https://nid.naver.com/oauth2.0/token",
		},
	}

	client = h.naverConfig.Client(context.Background(), &oauth2.Token{
		AccessToken: accessToken,
	})

	resp, err = client.Get("https://openapi.naver.com/v1/nid/me")
	if err != nil || resp.StatusCode != http.StatusOK {
		return c.Redirect().To(redirectPath)
	}
	defer resp.Body.Close()

	var userInfo models.NaverUser
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return c.Redirect().To(redirectPath)
	}

	id := userInfo.Response.Email
	name := userInfo.Response.Nickname
	profile := userInfo.Response.ProfileImage
	userUid := h.UtilRegisterUser(id, name, profile)
	if userUid < 1 {
		return c.Redirect().To(redirectPath)
	}

	return h.UtilFinishLogin(c, userUid)
}

// 카카오 OAuth 로그인을 위해 리다이렉트
func (h *TsboardOAuth2Handler) KakaoOAuthRequestHandler(c fiber.Ctx) error {
	state := uuid.New().String()[:10]
	utils.SaveCookie(c, "tsboard_oauth_state", state, 1)

	h.kakaoConfig = oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s%s/goapi/auth/kakao/callback", configs.Env.URL, configs.Env.URLPrefix),
		ClientID:     configs.Env.OAuthKakaoID,
		ClientSecret: configs.Env.OAuthKakaoSecret,
		Scopes:       []string{"account_email", "profile_image", "profile_nickname"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://kauth.kakao.com/oauth/authorize",
			TokenURL: "https://kauth.kakao.com/oauth/token",
		},
	}

	url := h.kakaoConfig.AuthCodeURL(state)
	return c.Redirect().To(url)
}

// 카카오 OAuth 콜백 핸들러
func (h *TsboardOAuth2Handler) KakaoOAuthCallbackHandler(c fiber.Ctx) error {
	redirectPath := fmt.Sprintf("%s%s", configs.Env.URL, configs.Env.URLPrefix)
	if configs.Env.OAuthKakaoID == "" {
		return c.Redirect().To(redirectPath)
	}

	token, err := utils.OAuth2ExchangeToken(c, h.kakaoConfig)
	if err != nil {
		return c.Redirect().To(redirectPath)
	}

	client := h.kakaoConfig.Client(context.Background(), token)
	resp, err := client.Get("https://kapi.kakao.com/v2/user/me")
	if err != nil {
		return c.Redirect().To(redirectPath)
	}
	defer resp.Body.Close()

	var userInfo models.KakaoUser
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return c.Redirect().To(redirectPath)
	}

	id := userInfo.KakaoAccount.Email
	name := userInfo.KakaoAccount.Profile.Nickname
	profile := userInfo.KakaoAccount.Profile.ProfileImageUrl
	userUid := h.UtilRegisterUser(id, name, profile)
	if userUid < 1 {
		return c.Redirect().To(redirectPath)
	}
	return h.UtilFinishLogin(c, userUid)
}

// 쿠키에 저장해둔 회원 정보 내려받기
func (h *TsboardOAuth2Handler) RequestUserInfoHandler(c fiber.Ctx) error {
	myinfo := c.Cookies("tsboard_myinfo")
	if myinfo == "" {
		return utils.Err(c, "Unable to read your data from cookie", models.CODE_FAILED_OPERATION)
	}
	data, err := base64.URLEncoding.DecodeString(myinfo)
	if err != nil {
		return utils.Err(c, "Unable to decode data", models.CODE_FAILED_OPERATION)
	}

	var info models.MyInfoResult
	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		return utils.Err(c, "Unable to unmarshal", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, info)
}

// 이미 등록된 사용자인지 확인하고 필요 시 등록 후 고유번호 반환
func (h *TsboardOAuth2Handler) UtilRegisterUser(id string, name string, profile string) uint {
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
func (h *TsboardOAuth2Handler) UtilFinishLogin(c fiber.Ctx, userUid uint) error {
	auth, refresh := h.service.OAuth.GenerateTokens(userUid)
	h.service.OAuth.SaveRefreshToken(userUid, refresh)

	user := h.service.OAuth.GetUserInfo(userUid)
	user.Token = auth
	user.Refresh = refresh
	myinfo, err := utils.ConvJsonString(user)
	if err != nil {
		return err
	}
	utils.SaveCookie(c, "tsboard_myinfo", myinfo, 1)
	return c.Redirect().To(fmt.Sprintf("%s%s/login/oauth", configs.Env.URL, configs.Env.URLPrefix))
}

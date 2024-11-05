package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"golang.org/x/oauth2"
)

// 상태 검사 및 토큰 교환 후 토큰 반환
func exchangeToken(w http.ResponseWriter, r *http.Request, cfg *oauth2.Config) (*oauth2.Token, error) {
	cookie, err := r.Cookie("tsboard_oauth_state")
	if err != nil || cookie.Value != r.FormValue("state") {
		http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
		return nil, fmt.Errorf("wrong state: %v", err)
	}

	code := r.FormValue("code")
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
		return nil, fmt.Errorf("token exchange failed: %v", err)
	}
	return token, nil
}

// 이미 등록된 사용자인지 확인하고 필요 시 등록 후 고유번호 반환
func registerUser(s *services.Service, id string, name string, profile string) uint {
	isRegistered := s.Auth.CheckEmailExists(id)
	var userUid uint
	if !isRegistered {
		userUid = s.OAuth.RegisterOAuthUser(id, name, profile)
	} else {
		userUid = s.OAuth.GetUserUid(id)
	}
	return userUid
}

// 토큰 저장 및 쿠키에 사용자 정보 전달
func finishOAuthLogin(s *services.Service, w http.ResponseWriter, r *http.Request, userUid uint) {
	auth, refresh := s.OAuth.GenerateTokens(userUid)
	s.OAuth.SaveRefreshToken(userUid, refresh)

	user := s.OAuth.GetUserInfo(userUid)
	user.Token = auth
	user.Refresh = refresh
	myinfo, err := utils.ConvertJsonString(user)
	if err != nil {
		return
	}
	utils.SaveCookie(w, "tsboard_myinfo", myinfo, 1)
	http.Redirect(w, r, fmt.Sprintf("%s/login/oauth", configs.Env.URL), http.StatusTemporaryRedirect)
}

// /////////////////////////////////////
// 구글 OAuth 로그인을 위해 리다이렉트
// /////////////////////////////////////
func GoogleOAuthRequestHandler(s *services.Service, cfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := uuid.New().String()[:10]
		utils.SaveCookie(w, "tsboard_oauth_state", state, 1)

		url := cfg.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// 구글 OAuth 콜백 핸들러
func GoogleOAuthCallbackHandler(s *services.Service, cfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if configs.Env.OAuthGoogleID == "" {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		token, err := exchangeToken(w, r, cfg)
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		client := cfg.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		var userInfo models.GoogleUser
		err = json.NewDecoder(resp.Body).Decode(&userInfo)
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		id := userInfo.Email
		name := userInfo.Name
		profile := userInfo.Picture
		userUid := registerUser(s, id, name, profile)
		if userUid < 1 {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		finishOAuthLogin(s, w, r, userUid)
	}
}

var naverRedirectURL string

// ///////////////////////////////////////
// 네이버 OAuth 로그인을 위해 리다이렉트
// ///////////////////////////////////////
func NaverOAuthRequestHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := uuid.New().String()[:10]
		utils.SaveCookie(w, "tsboard_oauth_state", state, 1)

		naverRedirectURL = fmt.Sprintf("%s/goapi/auth/naver/callback", configs.Env.URL)
		url := fmt.Sprintf("https://nid.naver.com/oauth2.0/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", configs.Env.OAuthNaverID, naverRedirectURL, state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// 네이버 OAuth 콜백 핸들러
func NaverOAuthCallbackHandler(s *services.Service, cfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if configs.Env.OAuthNaverID == "" {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		code := r.FormValue("code")
		state := r.FormValue("state")

		cookie, err := r.Cookie("tsboard_oauth_state")
		if err != nil || cookie.Value != state {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		apiURL := fmt.Sprintf(
			"https://nid.naver.com/oauth2.0/token?grant_type=authorization_code&client_id=%s&client_secret=%s&redirect_uri=%s&code=%s&state=%s",
			configs.Env.OAuthNaverID, configs.Env.OAuthNaverSecret, url.QueryEscape(naverRedirectURL), code, state,
		)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		req.Header.Set("X-Naver-Client-Id", configs.Env.OAuthNaverID)
		req.Header.Set("X-Naver-Client-Secret", configs.Env.OAuthNaverSecret)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		var tokenResp map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		accessToken, ok := tokenResp["access_token"].(string)
		if !ok || accessToken == "" {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		client = cfg.Client(context.Background(), &oauth2.Token{
			AccessToken: accessToken,
		})

		resp, err = client.Get("https://openapi.naver.com/v1/nid/me")
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		var userInfo models.NaverUser
		err = json.NewDecoder(resp.Body).Decode(&userInfo)
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		id := userInfo.Response.Email
		name := userInfo.Response.Nickname
		profile := userInfo.Response.ProfileImage
		userUid := registerUser(s, id, name, profile)
		if userUid < 1 {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		finishOAuthLogin(s, w, r, userUid)
	}
}

// //////////////////////////////////////
// 카카오 OAuth 로그인을 위해 리다이렉트
// //////////////////////////////////////
func KakaoOAuthRequestHandler(s *services.Service, cfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := uuid.New().String()[:10]
		utils.SaveCookie(w, "tsboard_oauth_state", state, 1)

		url := cfg.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// 카카오 OAuth 콜백 핸들러
func KakaoOAuthCallbackHandler(s *services.Service, cfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if configs.Env.OAuthKakaoID == "" {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		token, err := exchangeToken(w, r, cfg)
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		client := cfg.Client(context.Background(), token)
		resp, err := client.Get("https://kapi.kakao.com/v2/user/me")
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		var userInfo models.KakaoUser
		err = json.NewDecoder(resp.Body).Decode(&userInfo)
		if err != nil {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}

		id := userInfo.KakaoAccount.Email
		name := userInfo.KakaoAccount.Profile.Nickname
		profile := userInfo.KakaoAccount.Profile.ProfileImageUrl
		userUid := registerUser(s, id, name, profile)
		if userUid < 1 {
			http.Redirect(w, r, configs.Env.URL, http.StatusTemporaryRedirect)
			return
		}
		finishOAuthLogin(s, w, r, userUid)
	}
}

// 쿠키에 저장해둔 회원 정보 내려받기
func RequestUserInfoHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		myinfo, err := r.Cookie("tsboard_myinfo")
		if err != nil {
			utils.ResponseError(w, "Unable to read your data from cookie")
			return
		}
		data, err := base64.URLEncoding.DecodeString(myinfo.Value)
		if err != nil {
			utils.ResponseError(w, "Unable to decode data")
			return
		}

		var info models.MyInfoResult
		err = json.Unmarshal([]byte(data), &info)
		if err != nil {
			utils.ResponseError(w, "Unable to unmarshal")
			return
		}
		utils.ResponseSuccess(w, info)
	}
}

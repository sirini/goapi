package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler interface {
	CheckEmailHandler(c fiber.Ctx) error
	CheckNameHandler(c fiber.Ctx) error
	LoadMyInfoHandler(c fiber.Ctx) error
	LogoutHandler(c fiber.Ctx) error
	ResetPasswordHandler(c fiber.Ctx) error
	RefreshAccessTokenHandler(c fiber.Ctx) error
	SigninHandler(c fiber.Ctx) error
	SignupHandler(c fiber.Ctx) error
	VerifyCodeHandler(c fiber.Ctx) error
	UpdateMyInfoHandler(c fiber.Ctx) error
}

type NuboAuthHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboAuthHandler(service *services.Service) *NuboAuthHandler {
	return &NuboAuthHandler{service: service}
}

// (회원가입 시) 이메일 주소가 이미 등록되어 있는지 확인하기
func (h *NuboAuthHandler) CheckEmailHandler(c fiber.Ctx) error {
	param := models.CheckEmailParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	id := param.Email
	if len(id) < 6 {
		return utils.Err(c, "invalid email, too short", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Auth.CheckEmailExists(id)
	return utils.Ok(c, result)
}

// (회원가입 시) 이름이 이미 등록되어 있는지 확인하기
func (h *NuboAuthHandler) CheckNameHandler(c fiber.Ctx) error {
	param := models.CheckNameParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	name := param.Name
	if len(name) < 2 {
		return utils.Err(c, "invalid name, too short", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Auth.CheckNameExists(name, 0)
	return utils.Ok(c, result)
}

// 로그인 한 사용자의 정보 불러오기
func (h *NuboAuthHandler) LoadMyInfoHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	myinfo := h.service.Auth.GetMyInfo(uint(actionUserUid))
	if myinfo.Uid < 1 {
		return utils.Err(c, "Unable to load your information", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, myinfo)
}

// 로그아웃 처리하기
func (h *NuboAuthHandler) LogoutHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	h.service.Auth.Logout(uint(actionUserUid))

	c.ClearCookie(
		"nubo-auth-token",
		"nubo-auth-refresh",
	)

	return utils.Ok(c, nil)
}

// 비밀번호 초기화하기
func (h *NuboAuthHandler) ResetPasswordHandler(c fiber.Ctx) error {
	id := c.FormValue("email")
	if !utils.IsValidEmail(id) {
		return utils.Err(c, "Failed to reset password, invalid ID(email)", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Auth.ResetPassword(id, c.Hostname())
	if !result {
		return utils.Err(c, "Unable to reset password, internal error", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, &models.ResetPasswordResult{
		Sendmail: configs.Env.GmailAppPassword != "",
	})
}

// 사용자의 기존 (액세스) 토큰이 만료되었을 때, 리프레시 토큰 유효한지 보고 새로 발급
func (h *NuboAuthHandler) RefreshAccessTokenHandler(c fiber.Ctx) error {
	actionUserUid, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	refreshToken := c.Cookies("nubo-auth-refresh")
	if len(refreshToken) < 1 {
		return utils.Err(c, "Invalid refresh token", models.CODE_INVALID_PARAMETER)
	}

	if ok := h.service.Auth.CheckRefreshToken(uint(actionUserUid), refreshToken); !ok {
		return utils.Err(c, "Refresh token has expired, please sign in again", models.CODE_FAILED_OPERATION)
	}

	newAuthToken, _, err := h.service.Auth.SaveTokensInCookie(c, uint(actionUserUid))
	if err != nil {
		return utils.Err(c, "Failed to save tokens: "+err.Error(), models.CODE_FAILED_OPERATION)
	}

	return utils.Ok(c, newAuthToken)
}

// 로그인 하기
func (h *NuboAuthHandler) SigninHandler(c fiber.Ctx) error {
	form := models.SigninParam{}
	if err := c.Bind().Body(&form); err != nil {
		return utils.Err(c, "Unable to marshal given input", models.CODE_INVALID_PARAMETER)
	}

	id := form.ID
	pw := form.Password // 일반 문자열이어야 함

	if len(pw) < 1 || !utils.IsValidEmail(id) {
		return utils.Err(c, "Failed to sign in, invalid ID or password", models.CODE_INVALID_PARAMETER)
	}

	user, storedHash := h.service.Auth.GetUserAndHash(id)
	if user.Uid < 1 {
		return utils.Err(c, "Unable to get an information, invalid ID or password", models.CODE_FAILED_OPERATION)
	}

	if len(storedHash) == 60 && strings.HasPrefix(storedHash, "$2") { // NUBO 이후 암호화
		err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pw))
		if err != nil {
			return utils.Err(c, "Failed to sign in, invalid ID or password", models.CODE_INVALID_PARAMETER)
		}

	} else if len(storedHash) == 64 { // TSBOARD 시절 암호화
		oldHash := sha256.Sum256([]byte(pw))
		deprecatedHash := hex.EncodeToString(oldHash[:])

		if deprecatedHash != storedHash {
			return utils.Err(c, "Failed to sign in, invalid ID or password", models.CODE_INVALID_PARAMETER)
		}

		newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err == nil {
			h.service.Auth.ChangeHashForPassword(user.Uid, string(newBcryptHash))
		}
	} else {
		return utils.Err(c, "Failed to sign in, invalid ID or password", models.CODE_INVALID_PARAMETER)
	}

	authToken, refreshToken, err := h.service.Auth.SaveTokensInCookie(c, user.Uid)
	if err != nil {
		return utils.Err(c, "Failed to save tokens: "+err.Error(), models.CODE_FAILED_OPERATION)
	}

	user.Token = authToken
	user.Refresh = refreshToken
	return utils.Ok(c, user)
}

// 회원가입 하기
func (h *NuboAuthHandler) SignupHandler(c fiber.Ctx) error {
	param := models.SignupParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if !utils.IsValidEmail(param.ID) {
		return utils.Err(c, "invalid id, not an email address", models.CODE_INVALID_PARAMETER)
	}
	param.Hostname = c.Hostname()

	result, err := h.service.Auth.Signup(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 인증 완료하기
func (h *NuboAuthHandler) VerifyCodeHandler(c fiber.Ctx) error {
	param := models.VerifyParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if !utils.IsValidEmail(param.ID) {
		return utils.Err(c, "invalid id, not an email address", models.CODE_INVALID_PARAMETER)
	}
	if len(param.Name) < 2 {
		return utils.Err(c, "invalid name, too short", models.CODE_INVALID_PARAMETER)
	}
	if len(param.Code) != 6 {
		return utils.Err(c, "invalid code, wrong length", models.CODE_INVALID_PARAMETER)
	}
	result := h.service.Auth.VerifyEmail(param)
	return utils.Ok(c, result)
}

// 로그인 한 사용자 정보 업데이트
func (h *NuboAuthHandler) UpdateMyInfoHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	name := c.FormValue("name")
	signature := c.FormValue("signature")
	password := c.FormValue("password") // 일반 문자열이어야 함

	if len(name) < 2 || len(name) > 30 {
		return utils.Err(c, "invalid length of name (2~30 characters would be acceptable)", models.CODE_INVALID_PARAMETER)
	}
	if isDup := h.service.Auth.CheckNameExists(name, uint(actionUserUid)); isDup {
		return utils.Err(c, "duplicated name, please choose another one", models.CODE_DUPLICATED_VALUE)
	}
	userInfo, err := h.service.User.GetUserInfo(uint(actionUserUid))
	if err != nil {
		return utils.Err(c, "unable to find your information", models.CODE_FAILED_OPERATION)
	}

	if len(password) > 7 {
		newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err == nil {
			h.service.Auth.ChangeHashForPassword(uint(actionUserUid), string(newBcryptHash))
		}
	}

	header, _ := c.FormFile("profile")
	parameter := models.UpdateUserInfoParam{
		UserUid:    uint(actionUserUid),
		Name:       name,
		Signature:  signature,
		Password:   password,
		Profile:    header,
		OldProfile: userInfo.Profile,
	}
	h.service.User.ChangeUserInfo(parameter)
	return utils.Ok(c, nil)
}

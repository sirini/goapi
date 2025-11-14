package handlers

import (
	"html"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
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

type TsboardAuthHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardAuthHandler(service *services.Service) *TsboardAuthHandler {
	return &TsboardAuthHandler{service: service}
}

// (회원가입 시) 이메일 주소가 이미 등록되어 있는지 확인하기
func (h *TsboardAuthHandler) CheckEmailHandler(c fiber.Ctx) error {
	id := c.FormValue("email")
	if !utils.IsValidEmail(id) {
		return utils.Err(c, "Invalid email address", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Auth.CheckEmailExists(id)
	if result {
		return utils.Err(c, "Email address is already in use", models.CODE_DUPLICATED_VALUE)
	}
	return utils.Ok(c, nil)
}

// (회원가입 시) 이름이 이미 등록되어 있는지 확인하기
func (h *TsboardAuthHandler) CheckNameHandler(c fiber.Ctx) error {
	name := c.FormValue("name")
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Auth.CheckNameExists(name, 0)
	if result {
		return utils.Err(c, "Name is already in use", models.CODE_DUPLICATED_VALUE)
	}
	return utils.Ok(c, nil)
}

// 로그인 한 사용자의 정보 불러오기
func (h *TsboardAuthHandler) LoadMyInfoHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	myinfo := h.service.Auth.GetMyInfo(uint(actionUserUid))
	if myinfo.Uid < 1 {
		return utils.Err(c, "Unable to load your information", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, myinfo)
}

// 로그아웃 처리하기
func (h *TsboardAuthHandler) LogoutHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	h.service.Auth.Logout(uint(actionUserUid))

	c.ClearCookie(
		"nubo-oauth-access",
		"nubo-oauth-refresh",
		"auth-token",
		"auth-refresh",
	)

	return utils.Ok(c, nil)
}

// 비밀번호 초기화하기
func (h *TsboardAuthHandler) ResetPasswordHandler(c fiber.Ctx) error {
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
func (h *TsboardAuthHandler) RefreshAccessTokenHandler(c fiber.Ctx) error {
	actionUserUid, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	refreshToken := c.FormValue("refresh")
	if len(refreshToken) < 1 {
		return utils.Err(c, "Invalid refresh token", models.CODE_INVALID_PARAMETER)
	}

	newAccessToken, ok := h.service.Auth.GetUpdatedAccessToken(uint(actionUserUid), refreshToken)
	if !ok {
		return utils.Err(c, "Refresh token has been expired", models.CODE_EXPIRED_TOKEN)
	}
	return utils.Ok(c, newAccessToken)
}

// 로그인 하기
func (h *TsboardAuthHandler) SigninHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	pw := c.FormValue("password")

	if len(pw) != 64 || !utils.IsValidEmail(id) {
		return utils.Err(c, "Failed to sign in, invalid ID or password", models.CODE_INVALID_PARAMETER)
	}

	user := h.service.Auth.Signin(id, pw)
	if user.Uid < 1 {
		return utils.Err(c, "Unable to get an information, invalid ID or password", models.CODE_FAILED_OPERATION)
	}

	return utils.Ok(c, user)
}

// 회원가입 하기
func (h *TsboardAuthHandler) SignupHandler(c fiber.Ctx) error {
	id := c.FormValue("email")
	pw := c.FormValue("password")
	name := c.FormValue("name")

	if len(pw) != 64 || !utils.IsValidEmail(id) {
		return utils.Err(c, "Failed to sign up, invalid ID or password", models.CODE_INVALID_PARAMETER)
	}

	result, err := h.service.Auth.Signup(models.SignupParameter{
		ID:       id,
		Password: pw,
		Name:     name,
		Hostname: c.Hostname(),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 인증 완료하기
func (h *TsboardAuthHandler) VerifyCodeHandler(c fiber.Ctx) error {
	targetStr := c.FormValue("target")
	code := c.FormValue("code")
	id := c.FormValue("email")
	pw := c.FormValue("password")
	name := c.FormValue("name")

	if len(pw) != 64 || !utils.IsValidEmail(id) {
		return utils.Err(c, "Failed to verify, invalid ID or password", models.CODE_INVALID_PARAMETER)
	}
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short", models.CODE_INVALID_PARAMETER)
	}
	if len(code) != 6 {
		return utils.Err(c, "Invalid code, wrong length", models.CODE_INVALID_PARAMETER)
	}
	target, err := strconv.ParseUint(targetStr, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	result := h.service.Auth.VerifyEmail(models.VerifyParameter{
		Target:   uint(target),
		Code:     code,
		Id:       id,
		Password: pw,
		Name:     name,
	})

	if !result {
		return utils.Err(c, "Failed to verify code", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 로그인 한 사용자 정보 업데이트
func (h *TsboardAuthHandler) UpdateMyInfoHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	name := html.EscapeString(c.FormValue("name"))
	signature := html.EscapeString(c.FormValue("signature"))
	password := c.FormValue("password")

	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short", models.CODE_INVALID_PARAMETER)
	}
	if isDup := h.service.Auth.CheckNameExists(name, uint(actionUserUid)); isDup {
		return utils.Err(c, "Duplicated name, please choose another one", models.CODE_DUPLICATED_VALUE)
	}
	userInfo, err := h.service.User.GetUserInfo(uint(actionUserUid))
	if err != nil {
		return utils.Err(c, "Unable to find your information", models.CODE_FAILED_OPERATION)
	}

	header, _ := c.FormFile("profile")
	parameter := models.UpdateUserInfoParameter{
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

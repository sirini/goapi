package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type UserHandler interface {
	ChangePasswordHandler(c fiber.Ctx) error
	CheckReportedUserHandler(c fiber.Ctx) error
	LoadUserInfoHandler(c fiber.Ctx) error
	LoadUserPermissionHandler(c fiber.Ctx) error
	ManageUserPermissionHandler(c fiber.Ctx) error
	ReportUserHandler(c fiber.Ctx) error
}

type NuboUserHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboUserHandler(service *services.Service) *NuboUserHandler {
	return &NuboUserHandler{service: service}
}

// 비밀번호 변경하기
func (h *NuboUserHandler) ChangePasswordHandler(c fiber.Ctx) error {
	userCode := c.FormValue("code")
	newPassword := c.FormValue("password")

	if len(userCode) != 6 || len(newPassword) != 64 {
		return utils.Err(c, "Failed to change your password, invalid inputs", models.CODE_INVALID_PARAMETER)
	}

	verifyUid, err := strconv.ParseUint(c.FormValue("target"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.ChangePassword(uint(verifyUid), userCode, newPassword)
	if !result {
		return utils.Err(c, "Unable to change your password, internal error", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 이미 신고한 사용자인지 확인하기
func (h *NuboUserHandler) CheckReportedUserHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	targetUserUid, err := strconv.ParseUint(c.Query("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.CheckReportStatus(uint(actionUserUid), uint(targetUserUid))
	return utils.Ok(c, result)
}

// 사용자 정보 열람
func (h *NuboUserHandler) LoadUserInfoHandler(c fiber.Ctx) error {
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	userInfo, err := h.service.User.GetUserInfo(uint(targetUserUid))
	if err != nil {
		return utils.Err(c, "User not found", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, userInfo)
}

// 사용자 권한 및 리포트 응답 가져오기
func (h *NuboUserHandler) LoadUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.GetUserPermission(uint(actionUserUid), uint(targetUserUid))
	return utils.Ok(c, result)
}

// 사용자 권한 수정하기
func (h *NuboUserHandler) ManageUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	writePost, err := strconv.ParseBool(c.FormValue("writePost"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	writeComment, err := strconv.ParseBool(c.FormValue("writeComment"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	sendChat, err := strconv.ParseBool(c.FormValue("sendChatMessage"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	sendReport, err := strconv.ParseBool(c.FormValue("sendReport"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	login, err := strconv.ParseBool(c.FormValue("login"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	response := c.FormValue("response")

	param := models.UserPermissionReportResult{
		UserPermissionResult: models.UserPermissionResult{
			WritePost:       writePost,
			WriteComment:    writeComment,
			SendChatMessage: sendChat,
			SendReport:      sendReport,
		},
		Login:    login,
		UserUid:  uint(targetUserUid),
		Response: response,
	}

	err = h.service.User.ChangeUserPermission(uint(actionUserUid), param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 사용자 신고하기
func (h *NuboUserHandler) ReportUserHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param := models.UserReportParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	param.ActionUserUid = uint(actionUserUid)
	param.Content = utils.Escape(param.Content)
	result := h.service.User.ReportTargetUser(param)
	if !result {
		return utils.Err(c, "You have no permission to report other user", models.CODE_NO_PERMISSION)
	}
	return utils.Ok(c, nil)
}

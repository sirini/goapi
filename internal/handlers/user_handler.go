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
	LoadUserInfoHandler(c fiber.Ctx) error
	LoadUserPermissionHandler(c fiber.Ctx) error
	ManageUserPermissionHandler(c fiber.Ctx) error
	ReportUserHandler(c fiber.Ctx) error
}

type TsboardUserHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardUserHandler(service *services.Service) *TsboardUserHandler {
	return &TsboardUserHandler{service: service}
}

// 비밀번호 변경하기
func (h *TsboardUserHandler) ChangePasswordHandler(c fiber.Ctx) error {
	userCode := c.FormValue("code")
	newPassword := c.FormValue("password")

	if len(userCode) != 6 || len(newPassword) != 64 {
		return utils.Err(c, "Failed to change your password, invalid inputs", models.CODE_INVALID_PARAMETER)
	}

	verifyUid, err := strconv.ParseUint(c.FormValue("target"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.ChangePassword(uint(verifyUid), userCode, newPassword)
	if !result {
		return utils.Err(c, "Unable to change your password, internal error", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 사용자 정보 열람
func (h *TsboardUserHandler) LoadUserInfoHandler(c fiber.Ctx) error {
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	userInfo, err := h.service.User.GetUserInfo(uint(targetUserUid))
	if err != nil {
		return utils.Err(c, "User not found", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, userInfo)
}

// 사용자 권한 및 리포트 응답 가져오기
func (h *TsboardUserHandler) LoadUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.GetUserPermission(uint(actionUserUid), uint(targetUserUid))
	return utils.Ok(c, result)
}

// 사용자 권한 수정하기
func (h *TsboardUserHandler) ManageUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	writePost, err := strconv.ParseBool(c.FormValue("writePost"))
	if err != nil {
		return utils.Err(c, "Invalid writePost, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}
	writeComment, err := strconv.ParseBool(c.FormValue("writeComment"))
	if err != nil {
		return utils.Err(c, "Invalid writeComment, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}
	sendChat, err := strconv.ParseBool(c.FormValue("sendChatMessage"))
	if err != nil {
		return utils.Err(c, "Invalid sendChatMessage, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}
	sendReport, err := strconv.ParseBool(c.FormValue("sendReport"))
	if err != nil {
		return utils.Err(c, "Invalid sendReport, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}
	login, err := strconv.ParseBool(c.FormValue("login"))
	if err != nil {
		return utils.Err(c, "Invalid login, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
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
func (h *TsboardUserHandler) ReportUserHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	content := c.FormValue("content")
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	checkedBlackList, err := strconv.ParseBool(c.FormValue("checkedBlackList"))
	if err != nil {
		return utils.Err(c, "Invalid checkedBlackList, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}
	result := h.service.User.ReportTargetUser(uint(actionUserUid), uint(targetUserUid), checkedBlackList, content)
	if !result {
		return utils.Err(c, "You have no permission to report other user", models.CODE_NO_PERMISSION)
	}
	return utils.Ok(c, nil)
}

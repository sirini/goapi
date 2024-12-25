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
		return utils.Err(c, "Failed to change your password, invalid inputs")
	}

	verifyUid, err := strconv.ParseUint(c.FormValue("target"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target, not a valid number")
	}

	result := h.service.User.ChangePassword(uint(verifyUid), userCode, newPassword)
	if !result {
		return utils.Err(c, "Unable to change your password, internal error")
	}
	return utils.Ok(c, nil)
}

// 사용자 정보 열람
func (h *TsboardUserHandler) LoadUserInfoHandler(c fiber.Ctx) error {
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid targetUserUid, not a valid number")
	}

	userInfo, err := h.service.User.GetUserInfo(uint(targetUserUid))
	if err != nil {
		return utils.Err(c, "User not found")
	}
	return utils.Ok(c, userInfo)
}

// 사용자 권한 및 리포트 응답 가져오기
func (h *TsboardUserHandler) LoadUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	if actionUserUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number")
	}

	result := h.service.User.GetUserPermission(actionUserUid, uint(targetUserUid))
	return utils.Ok(c, result)
}

// 사용자 권한 수정하기
func (h *TsboardUserHandler) ManageUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	if actionUserUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid parameter(userUid), not a valid number")
	}
	writePost, err := strconv.ParseBool(c.FormValue("writePost"))
	if err != nil {
		return utils.Err(c, "Invalid parameter(writePost), not a valid boolean")
	}
	writeComment, err := strconv.ParseBool(c.FormValue("writeComment"))
	if err != nil {
		return utils.Err(c, "Invalid parameter(writeComment), not a valid boolean")
	}
	sendChat, err := strconv.ParseBool(c.FormValue("sendChatMessage"))
	if err != nil {
		return utils.Err(c, "Invalid parameter(sendChatMessage), not a valid boolean")
	}
	sendReport, err := strconv.ParseBool(c.FormValue("sendReport"))
	if err != nil {
		return utils.Err(c, "Invalid parameter(sendReport), not a valid boolean")
	}
	login, err := strconv.ParseBool(c.FormValue("login"))
	if err != nil {
		return utils.Err(c, "Invalid parameter(login), not a valid boolean")
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

	err = h.service.User.ChangeUserPermission(actionUserUid, param)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 사용자 신고하기
func (h *TsboardUserHandler) ReportUserHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	content := c.FormValue("content")
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid targetUserUid, not a valid number")
	}
	checkedBlackList, err := strconv.ParseBool(c.FormValue("checkedBlackList"))
	if err != nil {
		return utils.Err(c, "Invalid checkedBlackList, it should be able to convert a bool type")
	}
	result := h.service.User.ReportTargetUser(actionUserUid, uint(targetUserUid), checkedBlackList, content)
	if !result {
		return utils.Err(c, "You have no permission to report other user")
	}
	return utils.Ok(c, nil)
}

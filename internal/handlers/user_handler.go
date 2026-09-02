package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type UserHandler interface {
	AcknowledgeAchievementsHandler(c fiber.Ctx) error
	BlockUserHandler(c fiber.Ctx) error
	ChangePasswordHandler(c fiber.Ctx) error
	CheckReportedUserHandler(c fiber.Ctx) error
	LoadUserInfoHandler(c fiber.Ctx) error
	LoadUnannouncedAchievementsHandler(c fiber.Ctx) error
	LoadUserPermissionHandler(c fiber.Ctx) error
	ManageUserPermissionHandler(c fiber.Ctx) error
	ReportUserHandler(c fiber.Ctx) error
	UnblockUserHandler(c fiber.Ctx) error
	DeleteAccountHandler(c fiber.Ctx) error
}

func (h *NuboUserHandler) LoadUnannouncedAchievementsHandler(c fiber.Ctx) error {
	userUid := uint(utils.ExtractUserUid(c.Get(models.AUTH_KEY)))
	badges, err := h.service.User.GetUnannouncedAchievements(userUid)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, badges)
}

func (h *NuboUserHandler) AcknowledgeAchievementsHandler(c fiber.Ctx) error {
	userUid := uint(utils.ExtractUserUid(c.Get(models.AUTH_KEY)))
	param := models.BadgeAcknowledgeParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if err := h.service.User.AcknowledgeAchievements(userUid, param.Keys); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	return utils.Ok(c, nil)
}

type NuboUserHandler struct {
	service *services.Service
}

func (h *NuboUserHandler) BlockUserHandler(c fiber.Ctx) error {
	return h.changeBlockStatus(c, true)
}

func (h *NuboUserHandler) UnblockUserHandler(c fiber.Ctx) error {
	return h.changeBlockStatus(c, false)
}

func (h *NuboUserHandler) changeBlockStatus(c fiber.Ctx, blocked bool) error {
	actionUserUid := uint(utils.ExtractUserUid(c.Get(models.AUTH_KEY)))
	param := models.UserTargetParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	var err error
	if blocked {
		err = h.service.User.BlockUser(actionUserUid, param.TargetUserUid)
	} else {
		err = h.service.User.UnblockUser(actionUserUid, param.TargetUserUid)
	}
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

func (h *NuboUserHandler) DeleteAccountHandler(c fiber.Ctx) error {
	userUid := uint(utils.ExtractUserUid(c.Get(models.AUTH_KEY)))
	param := models.UserDeleteAccountParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if err := h.service.User.DeleteAccount(userUid, param.Confirmation); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// services.Service 주입 받기
func NewNuboUserHandler(service *services.Service) *NuboUserHandler {
	return &NuboUserHandler{service: service}
}

// 비밀번호 변경하기
func (h *NuboUserHandler) ChangePasswordHandler(c fiber.Ctx) error {
	param := models.UserChangePasswordParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	userCode := param.Code
	verifyUid := param.Target
	newPassword := param.Password // 일반 문자열이어야 함

	if len(userCode) != 6 || len(newPassword) < 8 {
		return utils.Err(c, "failed to change your password, invalid inputs", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.ChangePassword(verifyUid, userCode, newPassword)
	return utils.Ok(c, result)
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
	targetUserUid, err := strconv.ParseUint(c.Query("targetUserUid"), 10, 32)
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
	targetUserUid, err := strconv.ParseUint(c.Query("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	result := h.service.User.GetUserPermission(uint(actionUserUid), uint(targetUserUid))
	return utils.Ok(c, result)
}

// 사용자 권한 수정하기
func (h *NuboUserHandler) ManageUserPermissionHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param := models.UserPermissionManageParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	if err := h.service.User.ChangeUserPermission(uint(actionUserUid), param); err != nil {
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

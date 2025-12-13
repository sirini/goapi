package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type NotiHandler interface {
	CheckedAllNotiHandler(c fiber.Ctx) error
	CheckedSingleNotiHandler(c fiber.Ctx) error
	LoadNotiListHandler(c fiber.Ctx) error
}

type NuboNotiHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboNotiHandler(service *services.Service) *NuboNotiHandler {
	return &NuboNotiHandler{service: service}
}

// 알림 모두 확인하기 처리
func (h *NuboNotiHandler) CheckedAllNotiHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	h.service.Noti.CheckedAllNoti(uint(actionUserUid))
	return utils.Ok(c, nil)
}

// 하나의 알림만 확인 처리하기
func (h *NuboNotiHandler) CheckedSingleNotiHandler(c fiber.Ctx) error {
	notiUid, err := strconv.ParseUint(c.Params("notiUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid noti uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	h.service.Noti.CheckedSingleNoti(uint(notiUid))
	return utils.Ok(c, nil)
}

// 알림 목록 가져오기
func (h *NuboNotiHandler) LoadNotiListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	notis, err := h.service.Noti.GetUserNoti(uint(actionUserUid), uint(limit))
	if err != nil {
		return utils.Err(c, "Failed to load your notifications", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, notis)
}

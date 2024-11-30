package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/utils"
)

type NotiHandler interface {
	CheckedAllNotiHandler(c fiber.Ctx) error
	LoadNotiListHandler(c fiber.Ctx) error
}

type TsboardNotiHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardNotiHandler(service *services.Service) *TsboardNotiHandler {
	return &TsboardNotiHandler{service: service}
}

// 알림 모두 확인하기 처리
func (h *TsboardNotiHandler) CheckedAllNotiHandler(c fiber.Ctx) error {
	userUid := utils.ExtractUserUid(c.Get("Authorization"))
	if userUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}

	h.service.Noti.CheckedAllNoti(userUid, 10)
	return utils.Ok(c, nil)
}

// 알림 목록 가져오기
func (h *TsboardNotiHandler) LoadNotiListHandler(c fiber.Ctx) error {
	userUid := utils.ExtractUserUid(c.Get("Authorization"))
	if userUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}

	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	notis, err := h.service.Noti.GetUserNoti(userUid, uint(limit))
	if err != nil {
		return utils.Err(c, "Failed to load your notifications")
	}
	return utils.Ok(c, notis)
}

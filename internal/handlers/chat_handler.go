package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type ChatHandler interface {
	LoadChatListHandler(c fiber.Ctx) error
	LoadChatHistoryHandler(c fiber.Ctx) error
	SaveChatHandler(c fiber.Ctx) error
}

type TsboardChatHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardChatHandler(service *services.Service) *TsboardChatHandler {
	return &TsboardChatHandler{service: service}
}

// 오고 간 쪽지들의 목록 가져오기
func (h *TsboardChatHandler) LoadChatListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	if actionUserUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}

	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	chatItems, err := h.service.Chat.GetChattingList(actionUserUid, uint(limit))
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, chatItems)
}

// 특정인과 나눈 최근 쪽지들의 내용 가져오기
func (h *TsboardChatHandler) LoadChatHistoryHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	if actionUserUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}

	targetUserUid, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid (target) user uid, not a valid number")
	}

	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	chatHistories, err := h.service.Chat.GetChattingHistory(actionUserUid, uint(targetUserUid), uint(limit))
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, chatHistories)
}

// 쪽지 내용 저장하기
func (h *TsboardChatHandler) SaveChatHandler(c fiber.Ctx) error {

	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	if actionUserUid < 1 {
		return utils.Err(c, "Unable to get an user uid from token")
	}

	message := c.FormValue("message")
	if len(message) < 2 {
		return utils.Err(c, "Your message is too short, aborted")
	}

	targetUserUid64, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid targetUserUid parameter")
	}

	message = utils.Escape(message)
	targetUserUid := uint(targetUserUid64)

	if isPerm := h.service.Auth.CheckUserPermission(actionUserUid, models.USER_ACTION_SEND_CHAT); !isPerm {
		return utils.Err(c, "You don't have permission to send a chat message")
	}

	insertId := h.service.Chat.SaveChatMessage(actionUserUid, targetUserUid, message)
	if insertId < 1 {
		return utils.Err(c, "Failed to send a message")
	}
	return utils.Ok(c, insertId)
}

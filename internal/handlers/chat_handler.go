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
	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	chatItems, err := h.service.Chat.GetChattingList(uint(actionUserUid), uint(limit))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, chatItems)
}

// 특정인과 나눈 최근 쪽지들의 내용 가져오기
func (h *TsboardChatHandler) LoadChatHistoryHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	chatHistories, err := h.service.Chat.GetChattingHistory(uint(actionUserUid), uint(targetUserUid), uint(limit))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, chatHistories)
}

// 쪽지 내용 저장하기
func (h *TsboardChatHandler) SaveChatHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	message := c.FormValue("message")
	if len(message) < 2 {
		return utils.Err(c, "Your message is too short, aborted", models.CODE_INVALID_PARAMETER)
	}

	targetUserUid64, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid parameter", models.CODE_INVALID_PARAMETER)
	}

	message = utils.Escape(message)
	targetUserUid := uint(targetUserUid64)

	if isPerm := h.service.Auth.CheckUserPermission(uint(actionUserUid), models.USER_ACTION_SEND_CHAT); !isPerm {
		return utils.Err(c, "You don't have permission to send a chat message", models.CODE_NO_PERMISSION)
	}

	insertId := h.service.Chat.SaveChatMessage(uint(actionUserUid), targetUserUid, message)
	if insertId < 1 {
		return utils.Err(c, "Failed to send a message", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

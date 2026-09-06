package handlers

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type ChatHandler interface {
	LoadChatListHandler(c fiber.Ctx) error
	LoadChatHistoryHandler(c fiber.Ctx) error
	SaveChatHandler(c fiber.Ctx) error
	ReadChatHandler(c fiber.Ctx) error
}

type NuboChatHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboChatHandler(service *services.Service) *NuboChatHandler {
	return &NuboChatHandler{service: service}
}

// 오고 간 쪽지들의 목록 가져오기
func (h *NuboChatHandler) LoadChatListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
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
func (h *NuboChatHandler) LoadChatHistoryHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	targetUserUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
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
func (h *NuboChatHandler) SaveChatHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	payload := models.ChatSendMessage{}
	if err := c.Bind().Body(&payload); err != nil {
		return utils.Err(c, "Invalid parameters", models.CODE_INVALID_PARAMETER)
	}

	message, valid := normalizeChatMessage(payload.Message)
	targetUserUid := payload.TargetUserUid
	if !valid || targetUserUid < 1 || targetUserUid == uint(actionUserUid) {
		return utils.Err(c, "Invalid message or target user", models.CODE_INVALID_PARAMETER)
	}

	if isPerm := h.service.Auth.CheckUserPermission(uint(actionUserUid), models.USER_ACTION_SEND_CHAT); !isPerm {
		return utils.Err(c, "You don't have permission to send a chat message", models.CODE_NO_PERMISSION)
	}

	insertId := h.service.Chat.SaveChatMessage(uint(actionUserUid), targetUserUid, message)
	if insertId < 1 {
		return utils.Err(c, "Failed to send a message", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

// 현재 화면에 표시된 상대방 쪽지를 읽음 처리한다.
func (h *NuboChatHandler) ReadChatHandler(c fiber.Ctx) error {
	actionUserUid := uint(utils.ExtractUserUid(c.Get(models.AUTH_KEY)))
	payload := models.ChatReadRequest{}
	if err := c.Bind().Body(&payload); err != nil {
		return utils.Err(c, "Invalid parameters", models.CODE_INVALID_PARAMETER)
	}
	if payload.TargetUserUid < 1 || payload.ThroughUid < 1 || payload.TargetUserUid == actionUserUid {
		return utils.Err(c, "Invalid target user or message", models.CODE_INVALID_PARAMETER)
	}

	result, err := h.service.Chat.MarkChatRead(actionUserUid, payload.TargetUserUid, payload.ThroughUid)
	if errors.Is(err, services.ErrChatBlocked) {
		return utils.Err(c, err.Error(), models.CODE_NO_PERMISSION)
	}
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 채팅 메시지 앞뒤 공백을 정리하고 데이터베이스에 저장할 수 있는 길이인지 확인한다.
func normalizeChatMessage(message string) (string, bool) {
	normalized := strings.TrimSpace(message)
	length := utf8.RuneCountInString(normalized)
	return normalized, length > 0 && length <= 2000
}

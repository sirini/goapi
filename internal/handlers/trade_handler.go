package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type TradeHandler interface {
	TradeListHandler(c fiber.Ctx) error
	TradeModifyHandler(c fiber.Ctx) error
	TradeViewHandler(c fiber.Ctx) error
	TradeWriteHandler(c fiber.Ctx) error
	UpdateStatusHandler(c fiber.Ctx) error
}

type NuboTradeHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboTradeHandler(service *services.Service) *NuboTradeHandler {
	return &NuboTradeHandler{service: service}
}

// 거래 목록 가져오기 핸들러
func (h *NuboTradeHandler) TradeListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	postUidStrs := strings.Split(c.FormValue("postUids"), ",")
	results := make([]models.TradeResult, 0)

	for _, uidStr := range postUidStrs {
		uid, err := strconv.ParseUint(uidStr, 10, 32)
		if err != nil {
			return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
		}
		result, err := h.service.Trade.GetTradeItem(uint(uid), uint(actionUserUid))
		if err != nil {
			return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
		}
		results = append(results, result)
	}
	return utils.Ok(c, results)
}

// 거래 내용 수정하기 핸들러
func (h *NuboTradeHandler) TradeModifyHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckTradeWriteParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Trade.ModifyPost(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 거래 보기 핸들러
func (h *NuboTradeHandler) TradeViewHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	info, err := h.service.Trade.GetTradeItem(uint(postUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, info)
}

// 새 거래 작성하기 핸들러
func (h *NuboTradeHandler) TradeWriteHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckTradeWriteParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Trade.WritePost(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 거래 상태 변경 핸들러
func (h *NuboTradeHandler) UpdateStatusHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	newStatus, err := strconv.ParseUint(c.FormValue("newStatus"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Trade.UpdateStatus(uint(postUid), uint(newStatus), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

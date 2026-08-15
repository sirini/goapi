package handlers

import (
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type TradeHandler interface {
	TradeListHandler(c fiber.Ctx) error
	TradeLoadPostHandler(c fiber.Ctx) error
	TradeModifyHandler(c fiber.Ctx) error
	TradeViewHandler(c fiber.Ctx) error
	TradeWriteHandler(c fiber.Ctx) error
	UpdateStatusHandler(c fiber.Ctx) error
}

type NuboTradeHandler struct{ service *services.Service }

func NewNuboTradeHandler(service *services.Service) *NuboTradeHandler {
	return &NuboTradeHandler{service: service}
}

func (h *NuboTradeHandler) TradeListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	option, err := strconv.ParseUint(c.Query("option"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid search option", models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.Query("keyword"))
	if err != nil {
		return utils.Err(c, "invalid keyword", models.CODE_INVALID_PARAMETER)
	}
	page, err := strconv.ParseUint(c.Query("page"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid page", models.CODE_INVALID_PARAMETER)
	}
	boardUid := h.service.Board.GetBoardUid(c.Query("id"))
	config := h.service.Board.GetBoardConfig(boardUid)
	result, err := h.service.Trade.GetList(models.BoardListParam{
		BoardUid: boardUid, UserUid: uint(actionUserUid), Option: models.Search(option),
		Keyword: utils.Escape(keyword), Page: uint(page), Limit: config.RowCount,
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

func (h *NuboTradeHandler) TradeViewHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param := models.BoardViewParam{}
	if err := c.Bind().Query(&param); err != nil {
		return utils.Err(c, "invalid parameters", models.CODE_INVALID_PARAMETER)
	}
	param.UserUid = uint(max(actionUserUid, 0))
	param.BoardUid = h.service.Board.GetBoardUid(param.Id)
	if param.BoardUid < 1 {
		return utils.Err(c, "invalid board id", models.CODE_INVALID_PARAMETER)
	}
	result, err := h.service.Trade.GetView(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

func (h *NuboTradeHandler) TradeLoadPostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.Query("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid board uid", models.CODE_INVALID_PARAMETER)
	}
	postUid, err := strconv.ParseUint(c.Query("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid post uid", models.CODE_INVALID_PARAMETER)
	}
	result, err := h.service.Trade.LoadPost(uint(boardUid), uint(postUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

func (h *NuboTradeHandler) TradeWriteHandler(c fiber.Ctx) error {
	post, err := utils.CheckWriteParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	trade, err := utils.CheckTradeParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	result, err := h.service.Trade.WriteTradePost(models.TradeWriteParam{EditorWriteParam: post, TradeCommonItem: trade})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

func (h *NuboTradeHandler) TradeModifyHandler(c fiber.Ctx) error {
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid post uid", models.CODE_INVALID_PARAMETER)
	}
	post, err := utils.CheckWriteParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	trade, err := utils.CheckTradeParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	err = h.service.Trade.ModifyTradePost(models.TradeModifyParam{
		EditorModifyParam: models.EditorModifyParam{EditorWriteParam: post, PostUid: uint(postUid)},
		TradeCommonItem:   trade,
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

func (h *NuboTradeHandler) UpdateStatusHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid board uid", models.CODE_INVALID_PARAMETER)
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "invalid post uid", models.CODE_INVALID_PARAMETER)
	}
	status, err := strconv.ParseUint(c.FormValue("status"), 10, 8)
	if err != nil {
		return utils.Err(c, "invalid trade status", models.CODE_INVALID_PARAMETER)
	}
	err = h.service.Trade.UpdateTradeStatus(models.TradeStatusParam{
		BoardUid: uint(boardUid), PostUid: uint(postUid), UserUid: uint(actionUserUid), Status: models.TradeStatus(status),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

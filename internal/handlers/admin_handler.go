package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/utils"
)

type AdminHandler interface {
	AddBoardCategoryHandler(c fiber.Ctx) error
	BoardGeneralLoadHandler(c fiber.Ctx) error
	ChangeBoardGroupHandler(c fiber.Ctx) error
	ChangeBoardInfoHandler(c fiber.Ctx) error
	ChangeBoardNameHandler(c fiber.Ctx) error
	ChangeBoardRowHandler(c fiber.Ctx) error
	ChangeBoardTypeHandler(c fiber.Ctx) error
	ChangeBoardWidthHandler(c fiber.Ctx) error
	RemoveBoardCategoryHandler(c fiber.Ctx) error
	UseBoardCategoryHandler(c fiber.Ctx) error
}

type TsboardAdminHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardAdminHandler(service *services.Service) *TsboardAdminHandler {
	return &TsboardAdminHandler{service: service}
}

// 게시판에 카테고리 추가하는 핸들러
func (h *TsboardAdminHandler) AddBoardCategoryHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	categoryName := c.FormValue("newCategory")
	if len(categoryName) < 2 {
		return utils.Err(c, "Invalid category name, too short")
	}

	insertId := h.service.Admin.AddBoardCategory(uint(uid), categoryName)
	return utils.Ok(c, insertId)
}

// 게시판 관리화면 > 일반 기존 내용 불러오는 핸들러
func (h *TsboardAdminHandler) BoardGeneralLoadHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board ID")
	}
	config := h.service.Board.GetBoardConfig(boardUid)
	if config.Uid < 1 {
		return utils.Err(c, "Failed to load configuration")
	}

	return utils.Ok(c, config)
}

// 게시판 소속 그룹 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardGroupHandler(c fiber.Ctx) error {
	groupUid := c.FormValue("groupUid")
	_, err := strconv.ParseUint(groupUid, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid group uid, not a valid number")
	}

	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "group_uid", groupUid)
	return utils.Ok(c, nil)
}

// 게시판 설명 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardInfoHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	info := utils.Escape(c.FormValue("newInfo"))
	if len(info) < 2 {
		return utils.Err(c, "Invalid info, too short")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "info", info)
	return utils.Ok(c, nil)
}

// 게시판 이름 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardNameHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	name := utils.Escape(c.FormValue("newName"))
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "name", name)
	return utils.Ok(c, nil)
}

// 게시판 출력 라인 수 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardRowHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	rowCount := c.FormValue("newRows")
	_, err = strconv.ParseUint(rowCount, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid row, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "row_count", rowCount)
	return utils.Ok(c, nil)
}

// 게시판 타입(블로그, 갤러리 등) 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardTypeHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	boardType := c.FormValue("newType")
	_, err = strconv.ParseUint(boardType, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid type, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "type", boardType)
	return utils.Ok(c, nil)
}

// 게시판 폭 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardWidthHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	width := c.FormValue("newWidth")
	_, err = strconv.ParseUint(width, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid width, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "width", width)
	return utils.Ok(c, nil)
}

// 게시판에 특정 카테고리 제거하기 핸들러
func (h *TsboardAdminHandler) RemoveBoardCategoryHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	catUid, err := strconv.ParseUint(c.FormValue("categoryUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid category uid, not a valid number")
	}

	h.service.Admin.RemoveBoardCategory(uint(uid), uint(catUid))
	return utils.Ok(c, nil)
}

// 게시판에서 카테고리 기능 사용 or 사용 해제하는 핸들러
func (h *TsboardAdminHandler) UseBoardCategoryHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	use := c.FormValue("useCategory")
	_, err = strconv.ParseBool(use)
	if err != nil {
		return utils.Err(c, "Invalid use category, it should be 0 or 1")
	}

	h.service.Admin.UpdateBoardSetting(uint(uid), "use_category", use)
	return utils.Ok(c, nil)
}

package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type AdminHandler interface {
	AddBoardCategoryHandler(c fiber.Ctx) error
	BoardGeneralLoadHandler(c fiber.Ctx) error
	BoardLevelLoadHandler(c fiber.Ctx) error
	BoardPointLoadHandler(c fiber.Ctx) error
	ChangeBoardAdminHandler(c fiber.Ctx) error
	ChangeBoardGroupHandler(c fiber.Ctx) error
	ChangeBoardInfoHandler(c fiber.Ctx) error
	ChangeBoardLevelHandler(c fiber.Ctx) error
	ChangeBoardNameHandler(c fiber.Ctx) error
	ChangeBoardPointHandler(c fiber.Ctx) error
	ChangeBoardRowHandler(c fiber.Ctx) error
	ChangeBoardTypeHandler(c fiber.Ctx) error
	ChangeBoardWidthHandler(c fiber.Ctx) error
	ChangeGroupAdminHandler(c fiber.Ctx) error
	CreateBoardHandler(c fiber.Ctx) error
	DashboardItemLoadHandler(c fiber.Ctx) error
	DashboardLatestLoadHandler(c fiber.Ctx) error
	DashboardStatisticLoadHandler(c fiber.Ctx) error
	GetAdminCandidatesHandler(c fiber.Ctx) error
	GetBoardListHandler(c fiber.Ctx) error
	GroupGeneralLoadHandler(c fiber.Ctx) error
	RemoveBoardCategoryHandler(c fiber.Ctx) error
	RemoveBoardHandler(c fiber.Ctx) error
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

// 게시판 권한 설정 가져오기 핸들러
func (h *TsboardAdminHandler) BoardLevelLoadHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board ID")
	}

	perm, err := h.service.Admin.GetBoardLevelPolicy(boardUid)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, perm)
}

// 게시판 포인트 설정 가져오는 핸들러
func (h *TsboardAdminHandler) BoardPointLoadHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board ID")
	}

	policy, err := h.service.Admin.GetBoardPointPolicy(boardUid)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, policy)
}

// 게시판 관리자 변경하는 핸들러
func (h *TsboardAdminHandler) ChangeBoardAdminHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	newAdminUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number")
	}

	err = h.service.Admin.ChangeBoardAdmin(uint(boardUid), uint(newAdminUid))
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 게시판 소속 그룹 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardGroupHandler(c fiber.Ctx) error {
	groupUid := c.FormValue("groupUid")
	_, err := strconv.ParseUint(groupUid, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid group uid, not a valid number")
	}

	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "group_uid", groupUid)
	return utils.Ok(c, nil)
}

// 게시판 설명 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardInfoHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	info := utils.Escape(c.FormValue("newInfo"))
	if len(info) < 2 {
		return utils.Err(c, "Invalid info, too short")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "info", info)
	return utils.Ok(c, nil)
}

// 게시판 레벨 제한 변경하는 핸들러
func (h *TsboardAdminHandler) ChangeBoardLevelHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	listLevel, err := strconv.ParseInt(c.FormValue("list"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid list level, not a valid number")
	}

	viewLevel, err := strconv.ParseInt(c.FormValue("view"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid view level, not a valid number")
	}

	writeLevel, err := strconv.ParseInt(c.FormValue("write"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid write level, not a valid number")
	}

	commentLevel, err := strconv.ParseInt(c.FormValue("comment"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid comment level, not a valid number")
	}

	downloadLevel, err := strconv.ParseInt(c.FormValue("download"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid download level, not a valid number")
	}

	level := models.BoardActionLevel{
		BoardActionPoint: models.BoardActionPoint{
			View:     int(viewLevel),
			Write:    int(writeLevel),
			Comment:  int(commentLevel),
			Download: int(downloadLevel),
		},
		List: int(listLevel),
	}

	err = h.service.Admin.ChangeBoardLevelPolicy(uint(boardUid), level)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 게시판 이름 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardNameHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	name := utils.Escape(c.FormValue("newName"))
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "name", name)
	return utils.Ok(c, nil)
}

// 게시판 포인트 설정 변경하는 핸들러
func (h *TsboardAdminHandler) ChangeBoardPointHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	viewPoint, err := strconv.ParseInt(c.FormValue("view"), 10, 32)
	if err != nil || viewPoint < -10000 || viewPoint > 10000 {
		return utils.Err(c, "Invalid view point, not a valid number")
	}

	writePoint, err := strconv.ParseInt(c.FormValue("write"), 10, 32)
	if err != nil || writePoint < -10000 || writePoint > 10000 {
		return utils.Err(c, "Invalid write point, not a valid number")
	}

	commentPoint, err := strconv.ParseInt(c.FormValue("comment"), 10, 32)
	if err != nil || commentPoint < -10000 || commentPoint > 10000 {
		return utils.Err(c, "Invalid comment point, not a valid number")
	}

	downloadPoint, err := strconv.ParseInt(c.FormValue("download"), 10, 32)
	if err != nil || downloadPoint < -10000 || downloadPoint > 10000 {
		return utils.Err(c, "Invalid download point, not a valid number")
	}

	point := models.BoardActionPoint{
		View:     int(viewPoint),
		Write:    int(writePoint),
		Comment:  int(commentPoint),
		Download: int(downloadPoint),
	}

	err = h.service.Admin.ChangeBoardPointPolicy(uint(boardUid), point)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 게시판 출력 라인 수 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardRowHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	rowCount := c.FormValue("newRows")
	_, err = strconv.ParseUint(rowCount, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid row, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "row_count", rowCount)
	return utils.Ok(c, nil)
}

// 게시판 타입(블로그, 갤러리 등) 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardTypeHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	boardType := c.FormValue("newType")
	_, err = strconv.ParseUint(boardType, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid type, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "type", boardType)
	return utils.Ok(c, nil)
}

// 게시판 폭 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeBoardWidthHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	width := c.FormValue("newWidth")
	_, err = strconv.ParseUint(width, 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid width, not a valid number")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "width", width)
	return utils.Ok(c, nil)
}

// 그룹 관리자 변경하기 핸들러
func (h *TsboardAdminHandler) ChangeGroupAdminHandler(c fiber.Ctx) error {
	groupUid, err := strconv.ParseUint(c.FormValue("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid group uid, not a valid number")
	}
	newAdminUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target user uid, not a valid number")
	}

	err = h.service.Admin.ChangeGroupAdmin(uint(groupUid), uint(newAdminUid))
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 게시판 생성하기 핸들러
func (h *TsboardAdminHandler) CreateBoardHandler(c fiber.Ctx) error {
	groupUid, err := strconv.ParseUint(c.FormValue("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid group uid, not a valid number")
	}
	newBoardId := c.FormValue("newId")
	if len(newBoardId) < 2 {
		return utils.Err(c, "Invalid id, too short")
	}

	result := h.service.Admin.CreateNewBoard(uint(groupUid), newBoardId)
	return utils.Ok(c, result)
}

// 대시보드에서 그룹,게시판,회원 목록들 불러오는 핸들러
func (h *TsboardAdminHandler) DashboardItemLoadHandler(c fiber.Ctx) error {
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	items := h.service.Admin.GetDashboardItems(uint(bunch))
	return utils.Ok(c, items)
}

// 대시보드에서 최근 글,댓글,신고 목록들 불러오는 핸들러
func (h *TsboardAdminHandler) DashboardLatestLoadHandler(c fiber.Ctx) error {
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	latests := h.service.Admin.GetDashboardLatests(uint(bunch))
	return utils.Ok(c, latests)
}

// 대시보드에서 통계 데이터 불러오는 핸들러
func (h *TsboardAdminHandler) DashboardStatisticLoadHandler(c fiber.Ctx) error {
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	statistics := h.service.Admin.GetDashboardStatistics(uint(bunch))
	return utils.Ok(c, statistics)
}

// 관리자 변경 시 후보군 출력하는 핸들러
func (h *TsboardAdminHandler) GetAdminCandidatesHandler(c fiber.Ctx) error {
	name := utils.Escape(c.FormValue("name"))
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short")
	}
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	candidates, err := h.service.Admin.GetBoardAdminCandidates(name, uint(bunch))
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, candidates)
}

// 게시판 아이디 중복 방지를 위해 입력된 아이디와 유사한 목록 출력하는 핸들러
func (h *TsboardAdminHandler) GetBoardListHandler(c fiber.Ctx) error {
	boardId := c.FormValue("id")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	list := h.service.Admin.GetExistBoardIds(boardId, uint(bunch))
	return utils.Ok(c, list)
}

// 그룹 설정 및 소속 게시판 목록 반환하는 핸들러
func (h *TsboardAdminHandler) GroupGeneralLoadHandler(c fiber.Ctx) error {
	groupId := c.FormValue("id")
	config := h.service.Admin.GetGroupConfig(groupId)
	return utils.Ok(c, config)
}

// 게시판에 특정 카테고리 제거하기 핸들러
func (h *TsboardAdminHandler) RemoveBoardCategoryHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	catUid, err := strconv.ParseUint(c.FormValue("categoryUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid category uid, not a valid number")
	}

	h.service.Admin.RemoveBoardCategory(uint(boardUid), uint(catUid))
	return utils.Ok(c, nil)
}

// 게시판 삭제하기 핸들러
func (h *TsboardAdminHandler) RemoveBoardHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	err = h.service.Admin.RemoveBoard(uint(boardUid))
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 게시판에서 카테고리 기능 사용 or 사용 해제하는 핸들러
func (h *TsboardAdminHandler) UseBoardCategoryHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	use := c.FormValue("useCategory")
	_, err = strconv.ParseBool(use)
	if err != nil {
		return utils.Err(c, "Invalid use category, it should be 0 or 1")
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "use_category", use)
	return utils.Ok(c, nil)
}

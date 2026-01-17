package handlers

import (
	"math"
	"net/url"
	"strconv"
	"strings"

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
	ChangeGroupIdHandler(c fiber.Ctx) error
	CreateBoardHandler(c fiber.Ctx) error
	CreateGroupHandler(c fiber.Ctx) error
	DashboardItemLoadHandler(c fiber.Ctx) error
	DashboardLatestLoadHandler(c fiber.Ctx) error
	DashboardStatisticLoadHandler(c fiber.Ctx) error
	GetAdminCandidatesHandler(c fiber.Ctx) error
	GroupGeneralLoadHandler(c fiber.Ctx) error
	GroupListLoadHandler(c fiber.Ctx) error
	LatestCommentLoadHandler(c fiber.Ctx) error
	LatestCommentSearchHandler(c fiber.Ctx) error
	LatestPostLoadHandler(c fiber.Ctx) error
	LatestPostSearchHandler(c fiber.Ctx) error
	RemoveBoardCategoryHandler(c fiber.Ctx) error
	RemoveBoardHandler(c fiber.Ctx) error
	RemoveCommentHandler(c fiber.Ctx) error
	RemovePostHandler(c fiber.Ctx) error
	RemoveGroupHandler(c fiber.Ctx) error
	ReportListLoadHandler(c fiber.Ctx) error
	ReportListSearchHandler(c fiber.Ctx) error
	ShowSimilarBoardIdHandler(c fiber.Ctx) error
	ShowSimilarGroupIdHandler(c fiber.Ctx) error
	UseBoardCategoryHandler(c fiber.Ctx) error
	UserInfoLoadHandler(c fiber.Ctx) error
	UserInfoModifyHandler(c fiber.Ctx) error
	UserListLoadHandler(c fiber.Ctx) error
}

type NuboAdminHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboAdminHandler(service *services.Service) *NuboAdminHandler {
	return &NuboAdminHandler{service: service}
}

// 게시판에 카테고리 추가하는 핸들러
func (h *NuboAdminHandler) AddBoardCategoryHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	categoryName := c.FormValue("newCategory")
	if len(categoryName) < 2 {
		return utils.Err(c, "invalid category name", models.CODE_INVALID_PARAMETER)
	}

	insertId := h.service.Admin.AddBoardCategory(uint(uid), categoryName)
	return utils.Ok(c, insertId)
}

// 게시판 관리화면 > 일반 기존 내용 불러오는 핸들러
func (h *NuboAdminHandler) BoardGeneralLoadHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board ID", models.CODE_INVALID_PARAMETER)
	}
	config := h.service.Board.GetBoardConfig(boardUid)
	if config.Uid < 1 {
		return utils.Err(c, "Failed to load configuration", models.CODE_FAILED_OPERATION)
	}

	pairs := make([]models.Pair, 0)
	groups := h.service.Admin.GetGroupList()
	for _, group := range groups {
		pair := models.Pair{
			Uid:  group.Uid,
			Name: group.Id,
		}
		pairs = append(pairs, pair)
	}

	return utils.Ok(c, models.AdminBoardResult{
		Config: config,
		Groups: pairs,
	})
}

// 게시판 권한 설정 가져오기 핸들러
func (h *NuboAdminHandler) BoardLevelLoadHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board ID", models.CODE_INVALID_PARAMETER)
	}

	perm, err := h.service.Admin.GetBoardLevelPolicy(boardUid)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, perm)
}

// 게시판 포인트 설정 가져오는 핸들러
func (h *NuboAdminHandler) BoardPointLoadHandler(c fiber.Ctx) error {
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board ID", models.CODE_INVALID_PARAMETER)
	}

	policy, err := h.service.Admin.GetBoardPointPolicy(boardUid)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, policy)
}

// 게시판 관리자 변경하는 핸들러
func (h *NuboAdminHandler) ChangeBoardAdminHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	newAdminUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Admin.ChangeBoardAdmin(uint(boardUid), uint(newAdminUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시판 소속 그룹 변경하기 핸들러
func (h *NuboAdminHandler) ChangeBoardGroupHandler(c fiber.Ctx) error {
	groupUid := c.FormValue("groupUid")
	_, err := strconv.ParseUint(groupUid, 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "group_uid", groupUid)
	return utils.Ok(c, nil)
}

// 게시판 설명 변경하기 핸들러
func (h *NuboAdminHandler) ChangeBoardInfoHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	info := utils.Escape(c.FormValue("newInfo"))
	if len(info) < 2 {
		return utils.Err(c, "Invalid info, too short", models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "info", info)
	return utils.Ok(c, nil)
}

// 게시판 레벨 제한 변경하는 핸들러
func (h *NuboAdminHandler) ChangeBoardLevelHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	listLevel, err := strconv.ParseInt(c.FormValue("list"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	viewLevel, err := strconv.ParseInt(c.FormValue("view"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	writeLevel, err := strconv.ParseInt(c.FormValue("write"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	commentLevel, err := strconv.ParseInt(c.FormValue("comment"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	downloadLevel, err := strconv.ParseInt(c.FormValue("download"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
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
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시판 이름 변경하기 핸들러
func (h *NuboAdminHandler) ChangeBoardNameHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	name := utils.Escape(c.FormValue("newName"))
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short", models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "name", name)
	return utils.Ok(c, nil)
}

// 게시판 포인트 설정 변경하는 핸들러
func (h *NuboAdminHandler) ChangeBoardPointHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	viewPoint, err := strconv.ParseInt(c.FormValue("view"), 10, 32)
	if err != nil || viewPoint < -10000 || viewPoint > 10000 {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	writePoint, err := strconv.ParseInt(c.FormValue("write"), 10, 32)
	if err != nil || writePoint < -10000 || writePoint > 10000 {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	commentPoint, err := strconv.ParseInt(c.FormValue("comment"), 10, 32)
	if err != nil || commentPoint < -10000 || commentPoint > 10000 {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	downloadPoint, err := strconv.ParseInt(c.FormValue("download"), 10, 32)
	if err != nil || downloadPoint < -10000 || downloadPoint > 10000 {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	point := models.BoardActionPoint{
		View:     int(viewPoint),
		Write:    int(writePoint),
		Comment:  int(commentPoint),
		Download: int(downloadPoint),
	}

	err = h.service.Admin.ChangeBoardPointPolicy(uint(boardUid), point)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시판 출력 라인 수 변경하기 핸들러
func (h *NuboAdminHandler) ChangeBoardRowHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	rowCount := c.FormValue("newRows")
	_, err = strconv.ParseUint(rowCount, 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "row_count", rowCount)
	return utils.Ok(c, nil)
}

// 게시판 타입(블로그, 갤러리 등) 변경하기 핸들러
func (h *NuboAdminHandler) ChangeBoardTypeHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	boardType := c.FormValue("newType")
	_, err = strconv.ParseUint(boardType, 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "type", boardType)
	return utils.Ok(c, nil)
}

// 게시판 폭 변경하기 핸들러
func (h *NuboAdminHandler) ChangeBoardWidthHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	width := c.FormValue("newWidth")
	_, err = strconv.ParseUint(width, 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "width", width)
	return utils.Ok(c, nil)
}

// 그룹 관리자 변경하기 핸들러
func (h *NuboAdminHandler) ChangeGroupAdminHandler(c fiber.Ctx) error {
	groupUid, err := strconv.ParseUint(c.FormValue("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	newAdminUid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Admin.ChangeGroupAdmin(uint(groupUid), uint(newAdminUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 그룹 ID 변경하기 핸들러
func (h *NuboAdminHandler) ChangeGroupIdHandler(c fiber.Ctx) error {
	groupUid, err := strconv.ParseUint(c.FormValue("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	newGroupId := c.FormValue("changeGroupId")

	err = h.service.Admin.ChangeGroupId(uint(groupUid), newGroupId)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시판 생성하기 핸들러
func (h *NuboAdminHandler) CreateBoardHandler(c fiber.Ctx) error {
	groupUid, err := strconv.ParseUint(c.FormValue("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	newBoardId := c.FormValue("newId")
	if len(newBoardId) < 2 {
		return utils.Err(c, "Invalid id, too short", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Admin.CreateNewBoard(uint(groupUid), newBoardId)
	return utils.Ok(c, result)
}

// 그룹 생성하기 핸들러
func (h *NuboAdminHandler) CreateGroupHandler(c fiber.Ctx) error {
	newGroupId := c.FormValue("newId")
	if len(newGroupId) < 2 {
		return utils.Err(c, "Invalid id, too short", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Admin.CreateNewGroup(newGroupId)
	return utils.Ok(c, result)
}

// 대시보드에서 그룹,게시판,회원 목록들 불러오는 핸들러
func (h *NuboAdminHandler) DashboardItemLoadHandler(c fiber.Ctx) error {
	bunch, err := strconv.ParseUint(c.Query("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	items := h.service.Admin.GetDashboardItems(uint(bunch))
	return utils.Ok(c, items)
}

// 대시보드에서 최근 글,댓글,신고 목록들 불러오는 핸들러
func (h *NuboAdminHandler) DashboardLatestLoadHandler(c fiber.Ctx) error {
	bunch, err := strconv.ParseUint(c.Query("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	latests := h.service.Admin.GetDashboardLatests(uint(bunch))
	return utils.Ok(c, latests)
}

// 대시보드에서 통계 데이터 불러오는 핸들러
func (h *NuboAdminHandler) DashboardStatisticLoadHandler(c fiber.Ctx) error {
	bunch, err := strconv.ParseUint(c.Query("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	statistics := h.service.Admin.GetDashboardStatistics(uint(bunch))
	return utils.Ok(c, statistics)
}

// 관리자 변경 시 후보군 출력하는 핸들러
func (h *NuboAdminHandler) GetAdminCandidatesHandler(c fiber.Ctx) error {
	name, err := url.QueryUnescape(c.FormValue("name"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	name = utils.Escape(name)
	if len(name) < 2 {
		return utils.Err(c, "Invalid name, too short", models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	candidates, err := h.service.Admin.GetBoardAdminCandidates(name, uint(bunch))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, candidates)
}

// 그룹 설정 및 소속 게시판 목록 반환하는 핸들러
func (h *NuboAdminHandler) GroupGeneralLoadHandler(c fiber.Ctx) error {
	groupId := c.FormValue("id")
	config := h.service.Admin.GetGroupConfig(groupId)
	boards := h.service.Admin.GetBoardList(config.Uid)

	return utils.Ok(c, models.AdminGroupListResult{
		Config: config,
		Boards: boards,
	})
}

// 그룹 목록 가져오는 핸들러
func (h *NuboAdminHandler) GroupListLoadHandler(c fiber.Ctx) error {
	list := h.service.Admin.GetGroupList()
	return utils.Ok(c, list)
}

// 최근 댓글 불러오는 핸들러
func (h *NuboAdminHandler) LatestCommentLoadHandler(c fiber.Ctx) error {
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	parameter := models.AdminLatestParam{
		Page:    uint(page),
		Bunch:   uint(bunch),
		MaxUid:  0,
		Option:  models.SEARCH_NONE,
		Keyword: "",
	}
	comments := h.service.Admin.GetCommentList(parameter)
	return utils.Ok(c, comments)
}

// 댓글 검색하는 핸들러
func (h *NuboAdminHandler) LatestCommentSearchHandler(c fiber.Ctx) error {
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	parameter := models.AdminLatestParam{
		Page:    uint(page),
		Bunch:   uint(bunch),
		MaxUid:  0,
		Option:  models.Search(option),
		Keyword: keyword,
	}
	comments := h.service.Admin.GetCommentList(parameter)
	return utils.Ok(c, comments)
}

// 최근 글 불러오는 핸들러
func (h *NuboAdminHandler) LatestPostLoadHandler(c fiber.Ctx) error {
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	posts := h.service.Admin.GetLatestPosts(uint(page), uint(bunch))
	return utils.Ok(c, posts)
}

// 게시글 검색하는 핸들러
func (h *NuboAdminHandler) LatestPostSearchHandler(c fiber.Ctx) error {
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	parameter := models.AdminLatestParam{
		Page:    uint(page),
		Bunch:   uint(bunch),
		MaxUid:  0,
		Option:  models.Search(option),
		Keyword: keyword,
	}
	result := h.service.Admin.GetSearchedPosts(parameter)
	return utils.Ok(c, result)
}

// 게시판에 특정 카테고리 제거하기 핸들러
func (h *NuboAdminHandler) RemoveBoardCategoryHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	catUid, err := strconv.ParseUint(c.FormValue("categoryUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.RemoveBoardCategory(uint(boardUid), uint(catUid))
	return utils.Ok(c, nil)
}

// 게시판 삭제하기 핸들러
func (h *NuboAdminHandler) RemoveBoardHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Admin.RemoveBoard(uint(boardUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 댓글 삭제하기 핸들러
func (h *NuboAdminHandler) RemoveCommentHandler(c fiber.Ctx) error {
	targets := strings.Split(c.FormValue("targets"), ",")
	for _, target := range targets {
		commentUid, err := strconv.ParseUint(target, 10, 32)
		if err != nil {
			return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
		}
		h.service.Admin.RemoveComment(uint(commentUid))
	}
	return utils.Ok(c, nil)
}

// 게시글 삭제하기 핸들러
func (h *NuboAdminHandler) RemovePostHandler(c fiber.Ctx) error {
	targets := strings.Split(c.FormValue("targets"), ",")
	for _, target := range targets {
		postUid, err := strconv.ParseUint(target, 10, 32)
		if err != nil {
			return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
		}
		h.service.Admin.RemovePost(uint(postUid))
	}
	return utils.Ok(c, nil)
}

// 그룹 삭제하기 핸들러
func (h *NuboAdminHandler) RemoveGroupHandler(c fiber.Ctx) error {
	groupUid, err := strconv.ParseUint(c.FormValue("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Admin.RemoveGroup(uint(groupUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 신고 목록 가져오기 핸들러
func (h *NuboAdminHandler) ReportListLoadHandler(c fiber.Ctx) error {
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	isSolved, err := strconv.ParseBool(c.FormValue("isSolved"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	reports := h.service.Admin.GetReportList(uint(page), uint(bunch), isSolved)
	return utils.Ok(c, reports)
}

// 신고 목록 검색하기 핸들러
func (h *NuboAdminHandler) ReportListSearchHandler(c fiber.Ctx) error {
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	isSolved, err := strconv.ParseBool(c.FormValue("isSolved"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	parameter := models.AdminReportParam{
		AdminLatestParam: models.AdminLatestParam{
			Page:    uint(page),
			Bunch:   uint(bunch),
			MaxUid:  0,
			Option:  models.Search(option),
			Keyword: keyword,
		},
		IsSolved: isSolved,
	}
	reports := h.service.Admin.GetSearchedReports(parameter)
	return utils.Ok(c, reports)
}

// 게시판 아이디 중복 방지를 위해 입력된 아이디와 유사한 목록 출력하는 핸들러
func (h *NuboAdminHandler) ShowSimilarBoardIdHandler(c fiber.Ctx) error {
	boardId := c.FormValue("id")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	list := h.service.Admin.GetExistBoardIds(boardId, uint(bunch))
	return utils.Ok(c, list)
}

// 그룹 아이디 중복 방지를 위해 입력된 아이디와 유사한 목록 출력하는 핸들러
func (h *NuboAdminHandler) ShowSimilarGroupIdHandler(c fiber.Ctx) error {
	groupId := c.FormValue("id")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	list := h.service.Admin.GetExistGroupIds(groupId, uint(bunch))
	return utils.Ok(c, list)
}

// 게시판에서 카테고리 기능 사용 or 사용 해제하는 핸들러
func (h *NuboAdminHandler) UseBoardCategoryHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	use := c.FormValue("useCategory")
	_, err = strconv.ParseBool(use)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Admin.UpdateBoardSetting(uint(boardUid), "use_category", use)
	return utils.Ok(c, nil)
}

// 사용자 정보 가져오는 핸들러
func (h *NuboAdminHandler) UserInfoLoadHandler(c fiber.Ctx) error {
	userUid, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	info := h.service.Admin.GetUserInfo(uint(userUid))
	return utils.Ok(c, info)
}

// 사용자 정보 수정하는 핸들러
func (h *NuboAdminHandler) UserInfoModifyHandler(c fiber.Ctx) error {
	userUid64, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	userUid := uint(userUid64)
	userInfo, err := h.service.User.GetUserInfo(userUid)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	level, err := strconv.ParseUint(c.FormValue("level"), 10, 32)
	if err != nil || level > 9 {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	point, err := strconv.ParseUint(c.FormValue("point"), 10, 32)
	if err != nil || point > math.MaxUint {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	name := utils.Escape(c.FormValue("name"))
	if isDup := h.service.Auth.CheckNameExists(name, userUid); isDup {
		return utils.Err(c, "Duplicated name, please choose another one", models.CODE_DUPLICATED_VALUE)
	}
	signature := utils.Escape(c.FormValue("signature"))
	password := c.FormValue("password")
	header, _ := c.FormFile("profile")

	parameter := models.UpdateUserInfoParam{
		UserUid:    userUid,
		Name:       name,
		Signature:  signature,
		Password:   password,
		Profile:    header,
		OldProfile: userInfo.Profile,
	}
	h.service.User.ChangeUserInfo(parameter)
	h.service.Admin.UpdateUserLevelPoint(userUid, uint(level), uint(point))
	return utils.Ok(c, nil)
}

// 사용자 목록 조회하는 핸들러
func (h *NuboAdminHandler) UserListLoadHandler(c fiber.Ctx) error {
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	isBlocked, err := strconv.ParseBool(c.FormValue("isBlocked"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	parameter := models.AdminUserParam{
		AdminLatestParam: models.AdminLatestParam{
			Page:    uint(page),
			Bunch:   uint(bunch),
			MaxUid:  0,
			Option:  models.Search(option),
			Keyword: keyword,
		},
		IsBlocked: isBlocked,
	}
	result := h.service.Admin.GetUserList(parameter)
	return utils.Ok(c, result)
}

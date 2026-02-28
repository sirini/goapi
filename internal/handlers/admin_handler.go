package handlers

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type AdminHandler interface {
	BoardGeneralLoadHandler(c fiber.Ctx) error
	ChangeGroupAdminHandler(c fiber.Ctx) error
	ChangeGroupIdHandler(c fiber.Ctx) error
	CreateBoardHandler(c fiber.Ctx) error
	CreateGroupHandler(c fiber.Ctx) error
	DashboardUploadUsageHandler(c fiber.Ctx) error
	DashboardItemLoadHandler(c fiber.Ctx) error
	DashboardStatisticLoadHandler(c fiber.Ctx) error
	GetAdminCandidatesHandler(c fiber.Ctx) error
	GroupGeneralLoadHandler(c fiber.Ctx) error
	GroupListLoadHandler(c fiber.Ctx) error
	LatestCommentSearchHandler(c fiber.Ctx) error
	LatestPostSearchHandler(c fiber.Ctx) error
	ModifyBoardHandler(c fiber.Ctx) error
	RemoveBoardHandler(c fiber.Ctx) error
	RemoveCommentHandler(c fiber.Ctx) error
	RemovePostHandler(c fiber.Ctx) error
	RemoveGroupHandler(c fiber.Ctx) error
	RemoveUserHandler(c fiber.Ctx) error
	ReportListSearchHandler(c fiber.Ctx) error
	ShowSimilarBoardIdHandler(c fiber.Ctx) error
	ShowSimilarGroupIdHandler(c fiber.Ctx) error
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

// 게시판 설정값 및 그룹 목록 불러오는 핸들러
func (h *NuboAdminHandler) BoardGeneralLoadHandler(c fiber.Ctx) error {
	id := c.Query("id")
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
	param := models.AdminGroupChangeParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	err := h.service.Admin.ChangeGroupId(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시판 생성하기 핸들러
func (h *NuboAdminHandler) CreateBoardHandler(c fiber.Ctx) error {
	param := models.AdminBoardCreateParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	boardUid, err := h.service.Admin.CreateNewBoard(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	return utils.Ok(c, boardUid)
}

// 그룹 생성하기 핸들러
func (h *NuboAdminHandler) CreateGroupHandler(c fiber.Ctx) error {
	param := models.AdminGroupCreateParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	result, err := h.service.Admin.CreateNewGroup(param.NewGroupId)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	return utils.Ok(c, result)
}

// 대시보드에서 업로드 폴더의 총 사용량 가져오는 핸들러
func (h *NuboAdminHandler) DashboardUploadUsageHandler(c fiber.Ctx) error {
	path := c.Query("path")
	if len(path) < 1 {
		return utils.Err(c, "invalid a path", models.CODE_INVALID_PARAMETER)
	}
	size := h.service.Admin.GetDashboardUploadUsage(path)
	return utils.Ok(c, size)
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
	groupId := c.Query("id")
	config := h.service.Admin.GetGroupConfig(groupId)
	boards, err := h.service.Admin.GetBoardList(config.Uid)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

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

// 댓글 검색하는 핸들러
func (h *NuboAdminHandler) LatestCommentSearchHandler(c fiber.Ctx) error {
	page, err := strconv.ParseUint(c.Query("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	limit, err := strconv.ParseUint(c.Query("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	option, err := strconv.ParseUint(c.Query("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.Query("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	param := models.AdminLatestParam{
		Page:    uint(page),
		Limit:   uint(limit),
		Option:  models.Search(option),
		Keyword: keyword,
	}
	comments := h.service.Admin.GetSearchedComments(param)
	return utils.Ok(c, comments)
}

// 게시글 검색하는 핸들러
func (h *NuboAdminHandler) LatestPostSearchHandler(c fiber.Ctx) error {
	option, err := strconv.ParseUint(c.Query("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	page, err := strconv.ParseUint(c.Query("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	limit, err := strconv.ParseUint(c.Query("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	param := models.AdminLatestParam{
		Page:    uint(page),
		Limit:   uint(limit),
		Option:  models.Search(option),
		Keyword: keyword,
	}
	result := h.service.Admin.GetSearchedPosts(param)
	return utils.Ok(c, result)
}

// 게시판 설정 수정하기 핸들러
func (h *NuboAdminHandler) ModifyBoardHandler(c fiber.Ctx) error {
	param := models.AdminBoardModifyParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if err := h.service.Admin.ModifyExistBoard(param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시판 삭제하기 핸들러
func (h *NuboAdminHandler) RemoveBoardHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.Query("boardUid"), 10, 32)
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
	groupUid, err := strconv.ParseUint(c.Query("groupUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	if err := h.service.Admin.RemoveGroup(uint(groupUid)); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 사용자 삭제하기 핸들러
func (h *NuboAdminHandler) RemoveUserHandler(c fiber.Ctx) error {
	userUid, err := strconv.ParseUint(c.Query("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	if err := h.service.Admin.RemoveUser(uint(userUid)); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 신고 목록 검색하기 핸들러
func (h *NuboAdminHandler) ReportListSearchHandler(c fiber.Ctx) error {
	param := models.AdminReportSearchParam{}
	if err := c.Bind().Query(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	reportParam := models.AdminReportParam{
		AdminLatestParam: models.AdminLatestParam{
			Page:    uint(param.Page),
			Limit:   uint(param.Limit),
			Option:  models.Search(param.Option),
			Keyword: utils.Escape(param.Keyword),
		},
		IsSolved: param.IsSolved,
	}
	reports := h.service.Admin.GetSearchedReports(reportParam)
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

// 사용자 정보 가져오는 핸들러
func (h *NuboAdminHandler) UserInfoLoadHandler(c fiber.Ctx) error {
	userUid, err := strconv.ParseUint(c.Query("userUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	info := h.service.Admin.GetUserInfo(uint(userUid))
	return utils.Ok(c, info)
}

// 사용자 정보 수정하는 핸들러
func (h *NuboAdminHandler) UserInfoModifyHandler(c fiber.Ctx) error {
	param := models.UpdateUserInfoParam{}
	if err := c.Bind().Form(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	userInfo, err := h.service.User.GetUserInfo(param.UserUid)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	param.Signature = utils.Escape(param.Signature)
	param.OldProfile = userInfo.Profile

	h.service.User.ChangeUserInfo(param)
	return utils.Ok(c, nil)
}

// 사용자 목록 조회하는 핸들러
func (h *NuboAdminHandler) UserListLoadHandler(c fiber.Ctx) error {
	param := models.AdminUserParam{}
	if err := c.Bind().Query(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	param.Keyword = utils.Escape(param.Keyword)

	result := h.service.Admin.GetUserList(param)
	return utils.Ok(c, result)
}

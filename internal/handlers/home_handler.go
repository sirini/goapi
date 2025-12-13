package handlers

import (
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type HomeHandler interface {
	ShowVersionHandler(c fiber.Ctx) error
	CountingVisitorHandler(c fiber.Ctx) error
	LoadSidebarLinkHandler(c fiber.Ctx) error
	LoadAllPostsHandler(c fiber.Ctx) error
	LoadPostsByIdHandler(c fiber.Ctx) error
}

type NuboHomeHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboHomeHandler(service *services.Service) *NuboHomeHandler {
	return &NuboHomeHandler{service: service}
}

// 메세지 출력 테스트용 핸들러
func (h *NuboHomeHandler) ShowVersionHandler(c fiber.Ctx) error {
	return utils.Ok(c, &models.HomeVisitResult{
		Success:         true,
		OfficialWebsite: "nubohub.org",
		Version:         configs.Env.Version,
		License:         "MIT",
		Github:          "github.com/sirini/nubo",
	})
}

// 방문자 조회수 올리기 핸들러
func (h *NuboHomeHandler) CountingVisitorHandler(c fiber.Ctx) error {
	userUid, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		userUid = 0
	}
	h.service.Home.AddVisitorLog(uint(userUid))
	return utils.Ok(c, nil)
}

// 홈화면의 게시판 링크들 가져오기 핸들러
func (h *NuboHomeHandler) LoadSidebarLinkHandler(c fiber.Ctx) error {
	links, err := h.service.Home.GetSidebarLinks()
	if err != nil {
		return utils.Err(c, "Unable to load group/board links", models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, links)
}

// 홈화면에서 모든 최근 게시글들 가져오기 (검색 지원) 핸들러
func (h *NuboHomeHandler) LoadAllPostsHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	sinceUid64, err := strconv.ParseUint(c.FormValue("sinceUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid since uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil || bunch < 1 || bunch > 100 {
		return utils.Err(c, "Invalid bunch, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid option, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, "Invalid keyword, failed to unescape", models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	sinceUid := uint(sinceUid64)
	if sinceUid < 1 {
		sinceUid = h.service.Board.GetMaxUid() + 1
	}
	parameter := models.HomePostParam{
		SinceUid: sinceUid,
		Bunch:    uint(bunch),
		Option:   models.Search(option),
		Keyword:  keyword,
		UserUid:  uint(actionUserUid),
		BoardUid: 0,
	}

	result, err := h.service.Home.GetLatestPosts(parameter)
	if err != nil {
		return utils.Err(c, "Failed to get latest posts", models.CODE_FAILED_OPERATION)
	}

	return utils.Ok(c, result)
}

// 홈화면에서 지정된 게시판 ID에 해당하는 최근 게시글들 가져오기 핸들러
func (h *NuboHomeHandler) LoadPostsByIdHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	id := c.FormValue("id")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil || bunch < 1 || bunch > 100 {
		return utils.Err(c, "Invalid limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board id, unable to find board", models.CODE_INVALID_PARAMETER)
	}

	parameter := models.HomePostParam{
		SinceUid: h.service.Board.GetMaxUid() + 1,
		Bunch:    uint(bunch),
		Option:   models.SEARCH_NONE,
		Keyword:  "",
		UserUid:  uint(actionUserUid),
		BoardUid: uint(boardUid),
	}
	items, err := h.service.Home.GetLatestPosts(parameter)
	if err != nil {
		return utils.Err(c, "Failed to get latest posts from specific board", models.CODE_FAILED_OPERATION)
	}

	config := h.service.Board.GetBoardConfig(boardUid)
	return utils.Ok(c, models.BoardHomePostResult{
		Items:  items,
		Config: config,
	})
}

package handlers

import (
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	texttemplate "text/template"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type HomeHandler interface {
	ShowVersionHandler(c fiber.Ctx) error
	CountingVisitorHandler(c fiber.Ctx) error
	LoadSidebarLinkHandler(c fiber.Ctx) error
	LoadAllPostsHandler(c fiber.Ctx) error
	LoadMainPageHandler(c fiber.Ctx) error
	LoadPostsByIdHandler(c fiber.Ctx) error
	LoadSitemapHandler(c fiber.Ctx) error
}

type TsboardHomeHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardHomeHandler(service *services.Service) *TsboardHomeHandler {
	return &TsboardHomeHandler{service: service}
}

// 메세지 출력 테스트용 핸들러
func (h *TsboardHomeHandler) ShowVersionHandler(c fiber.Ctx) error {
	return utils.Ok(c, &models.HomeVisitResult{
		Success:         true,
		OfficialWebsite: "tsboard.dev",
		Version:         configs.Env.Version,
		License:         "MIT",
		Github:          "github.com/sirini/goapi",
	})
}

// 방문자 조회수 올리기 핸들러
func (h *TsboardHomeHandler) CountingVisitorHandler(c fiber.Ctx) error {
	userUid, err := strconv.ParseUint(c.FormValue("userUid"), 10, 32)
	if err != nil {
		userUid = 0
	}
	h.service.Home.AddVisitorLog(uint(userUid))
	return utils.Ok(c, nil)
}

// 홈화면의 사이드바에 사용할 게시판 링크들 가져오기 핸들러
func (h *TsboardHomeHandler) LoadSidebarLinkHandler(c fiber.Ctx) error {
	links, err := h.service.Home.GetSidebarLinks()
	if err != nil {
		return utils.Err(c, "Unable to load group/board links")
	}
	return utils.Ok(c, links)
}

// 홈화면에서 모든 최근 게시글들 가져오기 (검색 지원) 핸들러
func (h *TsboardHomeHandler) LoadAllPostsHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	sinceUid64, err := strconv.ParseUint(c.FormValue("sinceUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid since uid, not a valid number")
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil || bunch < 1 || bunch > 100 {
		return utils.Err(c, "Invalid bunch, not a valid number")
	}
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid option, not a valid number")
	}
	keyword, err := url.QueryUnescape(c.FormValue("keyword"))
	if err != nil {
		return utils.Err(c, "Invalid keyword, failed to unescape")
	}
	keyword = utils.Escape(keyword)

	sinceUid := uint(sinceUid64)
	if sinceUid < 1 {
		sinceUid = h.service.Board.GetMaxUid() + 1
	}
	parameter := models.HomePostParameter{
		SinceUid: sinceUid,
		Bunch:    uint(bunch),
		Option:   models.Search(option),
		Keyword:  keyword,
		UserUid:  uint(actionUserUid),
		BoardUid: 0,
	}

	result, err := h.service.Home.GetLatestPosts(parameter)
	if err != nil {
		return utils.Err(c, "Failed to get latest posts")
	}

	return utils.Ok(c, result)
}

// 검색엔진을 위한 메인 페이지 가져오는 핸들러
func (h *TsboardHomeHandler) LoadMainPageHandler(c fiber.Ctx) error {
	main := models.HomeMainPage{}
	main.Version = configs.Env.Version
	main.PageTitle = configs.Env.Title
	main.PageUrl = configs.Env.URL

	articles, err := h.service.Home.LoadMainPage(50)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	main.Articles = articles

	c.Set("Content-Type", "text/html")
	tmpl, err := template.New("main").Parse(templates.MainPageBody)
	if err != nil {
		return utils.Err(c, err.Error())
	}

	err = tmpl.Execute(c.Response().BodyWriter(), main)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return nil
}

// 홈화면에서 지정된 게시판 ID에 해당하는 최근 게시글들 가져오기 핸들러
func (h *TsboardHomeHandler) LoadPostsByIdHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil || bunch < 1 || bunch > 100 {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board id, unable to find board")
	}

	parameter := models.HomePostParameter{
		SinceUid: h.service.Board.GetMaxUid() + 1,
		Bunch:    uint(bunch),
		Option:   models.SEARCH_NONE,
		Keyword:  "",
		UserUid:  uint(actionUserUid),
		BoardUid: uint(boardUid),
	}
	items, err := h.service.Home.GetLatestPosts(parameter)
	if err != nil {
		return utils.Err(c, "Failed to get latest posts from specific board")
	}

	config := h.service.Board.GetBoardConfig(boardUid)
	return utils.Ok(c, models.BoardHomePostResult{
		Items:  items,
		Config: config,
	})
}

// 사이트맵 xml 내용 반환하기 핸들러
func (h *TsboardHomeHandler) LoadSitemapHandler(c fiber.Ctx) error {
	urls := []models.HomeSitemapURL{
		{
			Loc:        fmt.Sprintf("%s/goapi/seo/main.html", configs.Env.URL),
			LastMod:    time.Now().Format("2006-01-02"),
			ChangeFreq: "daily",
			Priority:   "1.0",
		},
	}

	boards := h.service.Home.GetBoardIDsForSitemap()
	urls = append(urls, boards...)

	c.Set("Content-Type", "application/xml")
	tmpl, err := texttemplate.New("sitemap").Parse(templates.SitemapBody)
	if err != nil {
		return utils.Err(c, err.Error())
	}

	err = tmpl.Execute(c.Response().BodyWriter(), urls)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return nil
}

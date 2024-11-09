package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 메세지 출력 테스트
func ShowVersionHandler(w http.ResponseWriter, r *http.Request) {
	utils.ResponseSuccess(w, &models.HomeVisitResult{
		Success:         true,
		OfficialWebsite: "tsboard.dev",
		Version:         configs.Env.Version,
		License:         "MIT",
		Github:          "github.com/sirini/goapi",
	})
}

// 방문자 조회수 올리기
func CountingVisitorHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUid, err := strconv.ParseUint(r.FormValue("userUid"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid user uid, not a valid number")
			return
		}
		s.Home.AddVisitorLog(uint(userUid))
		utils.ResponseSuccess(w, nil)
	}
}

// 홈화면의 사이드바에 사용할 게시판 링크들 가져오기
func LoadSidebarLinkHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		links, err := s.Home.GetSidebarLinks()
		if err != nil {
			utils.ResponseError(w, "Unable to load group/board links")
			return
		}
		utils.ResponseSuccess(w, links)
	}
}

// 홈화면에서 모든 최근 게시글들 가져오기 (검색 지원)
func LoadAllPostsHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sinceUid, err := strconv.ParseUint(r.FormValue("sinceUid"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid since uid, not a valid number")
			return
		}
		bunch, err := strconv.ParseUint(r.FormValue("bunch"), 10, 32)
		if err != nil || bunch < 1 || bunch > 100 {
			utils.ResponseError(w, "Invalid bunch, not a valid number")
			return
		}
		option, err := strconv.ParseUint(r.FormValue("option"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid option, not a valid number")
			return
		}
		actionUserUid, err := strconv.ParseUint(r.FormValue("accessUserUid"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid user uid, not a valid number")
			return
		}
		keyword := utils.Escape(r.FormValue("keyword"))

		parameter := &models.BoardPostParameter{
			SinceUid: uint(sinceUid),
			Bunch:    uint(bunch),
			Option:   models.Search(option),
			Keyword:  keyword,
			UserUid:  uint(actionUserUid),
			BoardUid: 0,
		}
		result, err := s.Home.GetLatestPosts(parameter)
		if err != nil {
			utils.ResponseError(w, "Failed to get latest posts")
			return
		}
		utils.ResponseSuccess(w, result)
	}
}

// 홈화면에서 지정된 게시판 ID에 해당하는 최근 게시글들 가져오기
func LoadPostsByIdHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardId := r.FormValue("id")
		bunch, err := strconv.ParseUint(r.FormValue("limit"), 10, 32)
		if err != nil || bunch < 1 || bunch > 100 {
			utils.ResponseError(w, "Invalid limit, not a valid number")
			return
		}
		actionUserUid, err := strconv.ParseUint(r.FormValue("accessUserUid"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid user uid, not a valid number")
			return
		}

		boardUid := s.Board.GetBoardUid(boardId)
		if boardUid < 1 {
			utils.ResponseError(w, "Invalid board id, unable to find board")
			return
		}

		parameter := &models.BoardPostParameter{
			SinceUid: s.Board.GetMaxUid() + 1,
			Bunch:    uint(bunch),
			Option:   models.SEARCH_NONE,
			Keyword:  "",
			UserUid:  uint(actionUserUid),
			BoardUid: uint(boardUid),
		}
		result, err := s.Home.GetLatestPosts(parameter)
		if err != nil {
			utils.ResponseError(w, "Failed to get latest posts from specific board")
			return
		}
		utils.ResponseSuccess(w, result)
	}
}

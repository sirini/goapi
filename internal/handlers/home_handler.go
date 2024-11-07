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

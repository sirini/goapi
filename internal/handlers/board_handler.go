package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 게시글 목록 가져오기
func LoadBoardListHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.FindUserUidFromHeader(r)
		boardId := r.FormValue("id")
		keyword := r.FormValue("keyword")
		page, err := strconv.ParseUint(r.FormValue("page"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid page, not a valid number")
			return
		}
		paging, err := strconv.ParseInt(r.FormValue("pagingDirection"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid direction of paging, not a valid number")
			return
		}
		sinceUid64, err := strconv.ParseUint(r.FormValue("sinceUid"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid since uid, not a valid number")
			return
		}
		option, err := strconv.ParseUint(r.FormValue("option"), 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid option, not a valid number")
			return
		}

		parameter := &models.BoardListParameter{}
		parameter.SinceUid = uint(sinceUid64)
		if parameter.SinceUid < 1 {
			parameter.SinceUid = s.Board.GetMaxUid() + 1
		}
		parameter.BoardUid = s.Board.GetBoardUid(boardId)
		config := s.Board.GetBoardConfig(parameter.BoardUid)

		parameter.Bunch = config.RowCount
		parameter.Option = models.Search(option)
		parameter.Keyword = utils.Escape(keyword)
		parameter.UserUid = actionUserUid
		parameter.Page = uint(page)
		parameter.Direction = models.Paging(paging)

		result, err := s.Board.LoadListItem(parameter)
		if err != nil {
			utils.ResponseError(w, "Failed to load a list of content")
			return
		}
		utils.ResponseSuccess(w, result)
	}
}

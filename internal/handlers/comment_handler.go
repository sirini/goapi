package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 댓글 목록 가져오기 핸들러
func CommentListHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.FindUserUidFromHeader(r)
		id := r.FormValue("id")
		postUid, err := strconv.ParseUint(r.FormValue("postUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid post uid, not a valid number")
			return
		}
		page, err := strconv.ParseUint(r.FormValue("page"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid page, not a valid number")
			return
		}
		bunch, err := strconv.ParseUint(r.FormValue("bunch"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid bunch, not a valid number")
			return
		}
		sinceUid, err := strconv.ParseUint(r.FormValue("sinceUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid since uid, not a valid number")
			return
		}
		paging, err := strconv.ParseInt(r.FormValue("pagingDirection"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid direction of paging, not a valid number")
			return
		}

		boardUid := s.Board.GetBoardUid(id)
		result, err := s.Comment.LoadComments(models.CommentListParameter{
			BoardUid:  boardUid,
			PostUid:   uint(postUid),
			UserUid:   actionUserUid,
			Page:      uint(page),
			Bunch:     uint(bunch),
			SinceUid:  uint(sinceUid),
			Direction: models.Paging(paging),
		})
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, result)
	}
}

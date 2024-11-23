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
		result, err := s.Comment.LoadList(models.CommentListParameter{
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

// 댓글에 좋아요 누르기 핸들러
func LikeCommentHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}
		commentUid, err := strconv.ParseUint(r.FormValue("commentUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid comment uid, not a valid number")
			return
		}
		liked, err := strconv.ParseBool(r.FormValue("liked"))
		if err != nil {
			utils.Error(w, "Invalid liked, not a boolean type")
			return
		}

		s.Comment.Like(models.CommentLikeParameter{
			BoardUid:   uint(boardUid),
			CommentUid: uint(commentUid),
			UserUid:    actionUserUid,
			Liked:      liked,
		})
		utils.Success(w, nil)
	}
}

// 기존 댓글 내용 수정하기 핸들러
func ModifyCommentHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parameter, err := utils.CheckCommentParameters(r)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		commentUid, err := strconv.ParseUint(r.FormValue("modifyTargetUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid modify target uid, not a valid number")
			return
		}

		err = s.Comment.Modify(models.CommentModifyParameter{
			CommentWriteParameter: parameter,
			CommentUid:            uint(commentUid),
		})
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, nil)
	}
}

// 댓글 삭제하기 핸들러
func RemoveCommentHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}
		commentUid, err := strconv.ParseUint(r.FormValue("removeTargetUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid comment uid, not a valid number")
			return
		}

		err = s.Comment.Remove(uint(commentUid), uint(boardUid), actionUserUid)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, nil)
	}
}

// 기존 댓글에 답글 작성하기 핸들러
func ReplyCommentHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parameter, err := utils.CheckCommentParameters(r)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		replyTargetUid, err := strconv.ParseUint(r.FormValue("replyTargetUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid reply target uid, not a valid number")
			return
		}

		insertId, err := s.Comment.Reply(models.CommentReplyParameter{
			CommentWriteParameter: parameter,
			ReplyTargetUid:        uint(replyTargetUid),
		})
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, insertId)
	}
}

// 새 댓글 작성하기 핸들러
func WriteCommentHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parameter, err := utils.CheckCommentParameters(r)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}

		insertId, err := s.Comment.Write(parameter)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, insertId)
	}
}

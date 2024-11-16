package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 게시글 목록 가져오기 핸들러
func BoardListHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.FindUserUidFromHeader(r)
		boardId := r.FormValue("id")
		keyword := r.FormValue("keyword")
		page, err := strconv.ParseUint(r.FormValue("page"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid page, not a valid number")
			return
		}
		paging, err := strconv.ParseInt(r.FormValue("pagingDirection"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid direction of paging, not a valid number")
			return
		}
		sinceUid64, err := strconv.ParseUint(r.FormValue("sinceUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid since uid, not a valid number")
			return
		}
		option, err := strconv.ParseUint(r.FormValue("option"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid option, not a valid number")
			return
		}

		parameter := models.BoardListParameter{}
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

		result, err := s.Board.GetListItem(parameter)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, result)
	}
}

// 게시글 보기 핸들러
func BoardViewHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.FindUserUidFromHeader(r)
		id := r.FormValue("id")
		postUid, err := strconv.ParseUint(r.FormValue("postUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid post uid, not a valid number")
			return
		}
		updateHit, err := strconv.ParseUint(r.FormValue("needUpdateHit"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid need update hit, not a valid number")
			return
		}
		limit, err := strconv.ParseUint(r.FormValue("latestLimit"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid latest limit, not a valid number")
			return
		}
		boardUid := s.Board.GetBoardUid(id)
		if boardUid < 1 {
			utils.Error(w, "Invalid board id, cannot find a board")
			return
		}

		parameter := models.BoardViewParameter{
			BoardViewCommonParameter: models.BoardViewCommonParameter{
				BoardUid: boardUid,
				PostUid:  uint(postUid),
				UserUid:  actionUserUid,
			},
			UpdateHit: updateHit > 0,
			Limit:     uint(limit),
		}

		result, err := s.Board.GetViewItem(parameter)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, result)
	}
}

// 첨부파일 다운로드 핸들러
func DownloadHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}
		fileUid, err := strconv.ParseUint(r.FormValue("fileUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid post uid, not a valid number")
			return
		}
		result, err := s.Board.Download(uint(boardUid), uint(fileUid), actionUserUid)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, result)
	}
}

// 갤러리 리스트 핸들러
func GalleryListHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.FindUserUidFromHeader(r)
		boardId := r.FormValue("id")
		keyword := r.FormValue("keyword")
		sinceUid64, err := strconv.ParseUint(r.FormValue("sinceUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid since uid, not a valid number")
			return
		}
		option, err := strconv.ParseUint(r.FormValue("option"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid option, not a valid number")
			return
		}
		page, err := strconv.ParseUint(r.FormValue("page"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid page, not a valid number")
			return
		}
		paging, err := strconv.ParseInt(r.FormValue("pagingDirection"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid direction of paging, not a valid number")
			return
		}

		parameter := models.BoardListParameter{}
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

		result := s.Board.GetGalleryList(parameter)
		utils.Success(w, result)
	}
}

// 게시글 좋아하기 핸들러
func LikePostHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}
		postUid, err := strconv.ParseUint(r.FormValue("postUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid post uid, not a valid number")
			return
		}
		liked, err := strconv.ParseUint(r.FormValue("liked"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid liked value, not a valid number")
			return
		}

		parameter := models.BoardViewLikeParameter{
			BoardViewCommonParameter: models.BoardViewCommonParameter{
				BoardUid: uint(boardUid),
				PostUid:  uint(postUid),
				UserUid:  actionUserUid,
			},
			Liked: liked > 0,
		}

		s.Board.LikeThisPost(parameter)
		utils.Success(w, nil)
	}
}

// 게시글 이동 대상 목록 가져오는 핸들러
func ListForMoveHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}

		boards, err := s.Board.GetBoardList(uint(boardUid), actionUserUid)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, boards)
	}
}

// 게시글 이동하기 핸들러
func MovePostHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}
		targetBoardUid, err := strconv.ParseUint(r.FormValue("targetBoardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid target board uid, not a valid number")
			return
		}
		postUid, err := strconv.ParseUint(r.FormValue("postUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid post uid, not a valid number")
			return
		}

		s.Board.MovePost(models.BoardMovePostParameter{
			BoardViewCommonParameter: models.BoardViewCommonParameter{
				BoardUid: uint(boardUid),
				PostUid:  uint(postUid),
				UserUid:  actionUserUid,
			},
			TargetBoardUid: uint(targetBoardUid),
		})
		utils.Success(w, nil)
	}
}

// 게시글 삭제하기 핸들러
func RemovePostHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid board uid, not a valid number")
			return
		}
		postUid, err := strconv.ParseUint(r.FormValue("postUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid post uid, not a valid number")
			return
		}

		s.Board.RemovePost(uint(boardUid), uint(postUid), actionUserUid)
		utils.Success(w, nil)
	}
}

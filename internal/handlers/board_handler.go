package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardHandler interface {
	BoardListHandler(c fiber.Ctx) error
	BoardViewHandler(c fiber.Ctx) error
	DownloadHandler(c fiber.Ctx) error
	GalleryListHandler(c fiber.Ctx) error
	GalleryLoadPhotoHandler(c fiber.Ctx) error
	LikePostHandler(c fiber.Ctx) error
	ListForMoveHandler(c fiber.Ctx) error
	MovePostHandler(c fiber.Ctx) error
	RemovePostHandler(c fiber.Ctx) error
}

type TsboardBoardHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardBoardHandler(service *services.Service) *TsboardBoardHandler {
	return &TsboardBoardHandler{service: service}
}

// 게시글 목록 가져오기 핸들러
func (h *TsboardBoardHandler) BoardListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	keyword := c.FormValue("keyword")
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid page, not a valid number")
	}
	paging, err := strconv.ParseInt(c.FormValue("pagingDirection"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid direction of paging, not a valid number")
	}
	sinceUid64, err := strconv.ParseUint(c.FormValue("sinceUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid since uid, not a valid number")
	}
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid option, not a valid number")
	}

	parameter := models.BoardListParameter{}
	parameter.SinceUid = uint(sinceUid64)
	if parameter.SinceUid < 1 {
		parameter.SinceUid = h.service.Board.GetMaxUid() + 1
	}
	parameter.BoardUid = h.service.Board.GetBoardUid(id)
	config := h.service.Board.GetBoardConfig(parameter.BoardUid)

	parameter.Bunch = config.RowCount
	parameter.Option = models.Search(option)
	parameter.Keyword = utils.Escape(keyword)
	parameter.UserUid = actionUserUid
	parameter.Page = uint(page)
	parameter.Direction = models.Paging(paging)

	result, err := h.service.Board.GetListItem(parameter)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, result)
}

// 게시글 보기 핸들러
func (h *TsboardBoardHandler) BoardViewHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}
	updateHit, err := strconv.ParseUint(c.FormValue("needUpdateHit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid need update hit, not a valid number")
	}
	limit, err := strconv.ParseUint(c.FormValue("latestLimit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid latest limit, not a valid number")
	}
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board id, cannot find a board")
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

	result, err := h.service.Board.GetViewItem(parameter)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, result)
}

// 첨부파일 다운로드 핸들러
func (h *TsboardBoardHandler) DownloadHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	fileUid, err := strconv.ParseUint(c.FormValue("fileUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}
	result, err := h.service.Board.Download(uint(boardUid), uint(fileUid), actionUserUid)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, result)
}

// 갤러리 리스트 핸들러
func (h *TsboardBoardHandler) GalleryListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	keyword := c.FormValue("keyword")
	sinceUid64, err := strconv.ParseUint(c.FormValue("sinceUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid since uid, not a valid number")
	}
	option, err := strconv.ParseUint(c.FormValue("option"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid option, not a valid number")
	}
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid page, not a valid number")
	}
	paging, err := strconv.ParseInt(c.FormValue("pagingDirection"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid direction of paging, not a valid number")
	}

	parameter := models.BoardListParameter{}
	parameter.SinceUid = uint(sinceUid64)
	if parameter.SinceUid < 1 {
		parameter.SinceUid = h.service.Board.GetMaxUid() + 1
	}
	parameter.BoardUid = h.service.Board.GetBoardUid(id)
	config := h.service.Board.GetBoardConfig(parameter.BoardUid)

	parameter.Bunch = config.RowCount
	parameter.Option = models.Search(option)
	parameter.Keyword = utils.Escape(keyword)
	parameter.UserUid = actionUserUid
	parameter.Page = uint(page)
	parameter.Direction = models.Paging(paging)

	result := h.service.Board.GetGalleryList(parameter)
	return utils.Ok(c, result)
}

// 갤러리 사진 열람하기 핸들러
func (h *TsboardBoardHandler) GalleryLoadPhotoHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	postUid, err := strconv.ParseUint(c.FormValue("no"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}
	boardUid := h.service.Board.GetBoardUid(id)
	result, err := h.service.Board.GetGalleryPhotos(boardUid, uint(postUid), actionUserUid)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, result)
}

// 게시글 좋아하기 핸들러
func (h *TsboardBoardHandler) LikePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}
	liked, err := strconv.ParseUint(c.FormValue("liked"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid liked value, not a valid number")
	}

	parameter := models.BoardViewLikeParameter{
		BoardViewCommonParameter: models.BoardViewCommonParameter{
			BoardUid: uint(boardUid),
			PostUid:  uint(postUid),
			UserUid:  actionUserUid,
		},
		Liked: liked > 0,
	}

	h.service.Board.LikeThisPost(parameter)
	return utils.Ok(c, nil)
}

// 게시글 이동 대상 목록 가져오는 핸들러
func (h *TsboardBoardHandler) ListForMoveHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}

	boards, err := h.service.Board.GetBoardList(uint(boardUid), actionUserUid)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, boards)
}

// 게시글 이동하기 핸들러
func (h *TsboardBoardHandler) MovePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	targetBoardUid, err := strconv.ParseUint(c.FormValue("targetBoardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target board uid, not a valid number")
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}

	h.service.Board.MovePost(models.BoardMovePostParameter{
		BoardViewCommonParameter: models.BoardViewCommonParameter{
			BoardUid: uint(boardUid),
			PostUid:  uint(postUid),
			UserUid:  actionUserUid,
		},
		TargetBoardUid: uint(targetBoardUid),
	})
	return utils.Ok(c, nil)
}

// 게시글 삭제하기 핸들러
func (h *TsboardBoardHandler) RemovePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}

	h.service.Board.RemovePost(uint(boardUid), uint(postUid), actionUserUid)
	return utils.Ok(c, nil)
}

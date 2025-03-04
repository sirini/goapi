package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type CommentHandler interface {
	CommentListHandler(c fiber.Ctx) error
	LikeCommentHandler(c fiber.Ctx) error
	ModifyCommentHandler(c fiber.Ctx) error
	RemoveCommentHandler(c fiber.Ctx) error
	ReplyCommentHandler(c fiber.Ctx) error
	WriteCommentHandler(c fiber.Ctx) error
}

type TsboardCommentHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardCommentHandler(service *services.Service) *TsboardCommentHandler {
	return &TsboardCommentHandler{service: service}
}

// 댓글 목록 가져오기 핸들러
func (h *TsboardCommentHandler) CommentListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	id := c.FormValue("id")
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	page, err := strconv.ParseUint(c.FormValue("page"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid page, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid bunch, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	sinceUid, err := strconv.ParseUint(c.FormValue("sinceUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid since uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	paging, err := strconv.ParseInt(c.FormValue("pagingDirection"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid direction of paging, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	boardUid := h.service.Board.GetBoardUid(id)
	result, err := h.service.Comment.LoadList(models.CommentListParameter{
		BoardUid:  boardUid,
		PostUid:   uint(postUid),
		UserUid:   uint(actionUserUid),
		Page:      uint(page),
		Bunch:     uint(bunch),
		SinceUid:  uint(sinceUid),
		Direction: models.Paging(paging),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 댓글에 좋아요 누르기 핸들러
func (h *TsboardCommentHandler) LikeCommentHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	commentUid, err := strconv.ParseUint(c.FormValue("commentUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid comment uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	liked, err := strconv.ParseBool(c.FormValue("liked"))
	if err != nil {
		return utils.Err(c, "Invalid liked, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}

	h.service.Comment.Like(models.CommentLikeParameter{
		BoardUid:   uint(boardUid),
		CommentUid: uint(commentUid),
		UserUid:    uint(actionUserUid),
		Liked:      liked,
	})
	return utils.Ok(c, nil)
}

// 기존 댓글 내용 수정하기 핸들러
func (h *TsboardCommentHandler) ModifyCommentHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckCommentParameters(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	commentUid, err := strconv.ParseUint(c.FormValue("targetUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid modify target uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Comment.Modify(models.CommentModifyParameter{
		CommentWriteParameter: parameter,
		CommentUid:            uint(commentUid),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 댓글 삭제하기 핸들러
func (h *TsboardCommentHandler) RemoveCommentHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	commentUid, err := strconv.ParseUint(c.FormValue("removeTargetUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid comment uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Comment.Remove(uint(commentUid), uint(boardUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 기존 댓글에 답글 작성하기 핸들러
func (h *TsboardCommentHandler) ReplyCommentHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckCommentParameters(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	replyTargetUid, err := strconv.ParseUint(c.FormValue("targetUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid reply target uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	insertId, err := h.service.Comment.Reply(models.CommentReplyParameter{
		CommentWriteParameter: parameter,
		ReplyTargetUid:        uint(replyTargetUid),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

// 새 댓글 작성하기 핸들러
func (h *TsboardCommentHandler) WriteCommentHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckCommentParameters(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	insertId, err := h.service.Comment.Write(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

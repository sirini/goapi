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

type NuboCommentHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboCommentHandler(service *services.Service) *NuboCommentHandler {
	return &NuboCommentHandler{service: service}
}

// 댓글 목록 가져오기 핸들러
func (h *NuboCommentHandler) CommentListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param := models.CommentListParam{}
	if err := c.Bind().Query(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	param.UserUid = uint(actionUserUid)
	result, err := h.service.Comment.LoadList(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 댓글에 좋아요 누르기 핸들러
func (h *NuboCommentHandler) LikeCommentHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	commentUid, err := strconv.ParseUint(c.FormValue("commentUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid comment uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	liked, err := strconv.ParseBool(c.FormValue("liked"))
	if err != nil {
		return utils.Err(c, "Invalid liked, it should be 0 or 1", models.CODE_INVALID_PARAMETER)
	}

	h.service.Comment.Like(models.CommentLikeParam{
		BoardUid:   uint(boardUid),
		CommentUid: uint(commentUid),
		UserUid:    uint(actionUserUid),
		Liked:      liked,
	})
	return utils.Ok(c, nil)
}

// 기존 댓글 내용 수정하기 핸들러
func (h *NuboCommentHandler) ModifyCommentHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckCommentParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	commentUid, err := strconv.ParseUint(c.FormValue("targetUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid modify target uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	err = h.service.Comment.Modify(models.CommentModifyParam{
		CommentWriteParam: parameter,
		CommentUid:        uint(commentUid),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 댓글 삭제하기 핸들러
func (h *NuboCommentHandler) RemoveCommentHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
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
func (h *NuboCommentHandler) ReplyCommentHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckCommentParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	replyTargetUid, err := strconv.ParseUint(c.FormValue("targetUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid reply target uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	insertId, err := h.service.Comment.Reply(models.CommentReplyParam{
		CommentWriteParam: parameter,
		ReplyTargetUid:    uint(replyTargetUid),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

// 새 댓글 작성하기 핸들러
func (h *NuboCommentHandler) WriteCommentHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckCommentParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	insertId, err := h.service.Comment.Write(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

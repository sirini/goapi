package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type CommentHandler interface {
	CommentListHandler(c fiber.Ctx) error
	LikeCommentHandler(c fiber.Ctx) error
	SetReactionHandler(c fiber.Ctx) error
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
	if actionUserUid < 0 {
		actionUserUid = 0
	}
	param := models.CommentListParam{}
	if err := c.Bind().Query(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	param.UserUid = uint(actionUserUid)
	result, err := h.service.Comment.List(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 댓글에 좋아요 누르기 핸들러
func (h *NuboCommentHandler) LikeCommentHandler(c fiber.Ctx) error {
	param := models.CommentLikeParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param.UserUid = uint(actionUserUid)

	if err := h.service.Comment.Like(param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 댓글 다중 리액션 핸들러
func (h *NuboCommentHandler) SetReactionHandler(c fiber.Ctx) error {
	// reaction 필드의 누락과 명시적 null을 구분한다. 누락은 거부하고 null만 취소한다.
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(c.Body(), &fields); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if _, present := fields["reaction"]; !present {
		return utils.Err(c, "reaction field is required", models.CODE_INVALID_PARAMETER)
	}
	body := models.CommentReactionBody{}
	if err := c.Bind().Body(&body); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	if body.BoardUid == 0 || body.CommentUid == 0 {
		return utils.Err(c, "boardUid and commentUid are required", models.CODE_INVALID_PARAMETER)
	}
	reaction := ""
	if body.Reaction != nil {
		reaction = *body.Reaction
	}
	state, err := h.service.Comment.SetReaction(models.CommentReactionParam{
		BoardUid: body.BoardUid, CommentUid: body.CommentUid,
		UserUid:  uint(utils.ExtractUserUid(c.Get(models.AUTH_KEY))),
		Reaction: reaction, ReactionIsNull: body.Reaction == nil,
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, state)
}

// 기존 댓글 내용 수정하기 핸들러
func (h *NuboCommentHandler) ModifyCommentHandler(c fiber.Ctx) error {
	param := models.CommentModifyParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param.UserUid = uint(actionUserUid)

	if err := h.service.Comment.Modify(param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 댓글 삭제하기 핸들러
func (h *NuboCommentHandler) RemoveCommentHandler(c fiber.Ctx) error {
	param := models.CommentRemoveParam{}
	boardUid, err := strconv.ParseUint(c.Query("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	removeTargetUid, err := strconv.ParseUint(c.Query("removeTargetUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param.UserUid = uint(actionUserUid)
	param.BoardUid = uint(boardUid)
	param.RemoveTargetUid = uint(removeTargetUid)

	if err := h.service.Comment.Remove(param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 기존 댓글에 답글 작성하기 핸들러
func (h *NuboCommentHandler) ReplyCommentHandler(c fiber.Ctx) error {
	param := models.CommentReplyParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param.UserUid = uint(actionUserUid)
	param.Content = utils.Sanitize(param.Content)

	if len(param.Content) < 10 {
		return utils.Err(c, "content is too short", models.CODE_INVALID_PARAMETER)
	}

	insertId, err := h.service.Comment.Reply(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

// 새 댓글 작성하기 핸들러
func (h *NuboCommentHandler) WriteCommentHandler(c fiber.Ctx) error {
	param := models.CommentWriteParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param.UserUid = uint(actionUserUid)
	param.Content = utils.Sanitize(param.Content)

	if len(param.Content) < 10 {
		return utils.Err(c, "content is too short", models.CODE_INVALID_PARAMETER)
	}

	insertId, err := h.service.Comment.Write(param)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, insertId)
}

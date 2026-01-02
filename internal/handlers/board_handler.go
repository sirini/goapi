package handlers

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardHandler interface {
	BoardListHandler(c fiber.Ctx) error
	BoardRecentTagListHandler(c fiber.Ctx) error
	BoardViewHandler(c fiber.Ctx) error
	DownloadHandler(c fiber.Ctx) error
	LatestUserContentHandler(c fiber.Ctx) error
	LikePostHandler(c fiber.Ctx) error
	ListForMoveHandler(c fiber.Ctx) error
	MovePostHandler(c fiber.Ctx) error
	RemovePostHandler(c fiber.Ctx) error
	TransferHandler(c fiber.Ctx) error
}

// 다운로드 시 검증용으로 쓸 임시 토큰 구조체
type DownloadToken struct {
	Name   string
	Path   string
	Expiry time.Time
}

type NuboBoardHandler struct {
	service              *services.Service
	downloadTokenStorage map[string]DownloadToken
}

// services.Service 주입 받기
func NewNuboBoardHandler(service *services.Service) *NuboBoardHandler {
	return &NuboBoardHandler{service: service, downloadTokenStorage: make(map[string]DownloadToken)}
}

// 게시글 목록 가져오기 핸들러
func (h *NuboBoardHandler) BoardListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	id := c.Query("id")
	option, err := strconv.ParseUint(c.Query("option"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword, err := url.QueryUnescape(c.Query("keyword"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	keyword = utils.Escape(keyword)

	page, err := strconv.ParseUint(c.Query("page"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	parameter := models.BoardListParam{}
	parameter.BoardUid = h.service.Board.GetBoardUid(id)
	config := h.service.Board.GetBoardConfig(parameter.BoardUid)

	parameter.Limit = config.RowCount
	parameter.Option = models.Search(option)
	parameter.Keyword = keyword
	parameter.UserUid = uint(actionUserUid)
	parameter.Page = uint(page)

	result, err := h.service.Board.GetListItem(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 최근 사용된 해시태그 목록 보기 핸들러
func (h *NuboBoardHandler) BoardRecentTagListHandler(c fiber.Ctx) error {
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	result, err := h.service.Board.GetRecentTags(uint(boardUid), uint(limit))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 게시글 보기 핸들러
func (h *NuboBoardHandler) BoardViewHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	id := c.FormValue("id")
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	updateHit, err := strconv.ParseBool(c.FormValue("needUpdateHit"))
	if err != nil {
		return utils.Err(c, "Invalid need update hit, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	limit, err := strconv.ParseUint(c.FormValue("latestLimit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid latest limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	boardUid := h.service.Board.GetBoardUid(id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board id, cannot find a board", models.CODE_INVALID_PARAMETER)
	}

	parameter := models.BoardViewParam{
		BoardViewCommonParam: models.BoardViewCommonParam{
			BoardUid: boardUid,
			PostUid:  uint(postUid),
			UserUid:  uint(actionUserUid),
		},
		UpdateHit: updateHit,
		Limit:     uint(limit),
	}

	result, err := h.service.Board.GetViewItem(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 첨부파일 다운로드 핸들러
func (h *NuboBoardHandler) DownloadHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.Query("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	fileUid, err := strconv.ParseUint(c.Query("fileUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	result, err := h.service.Board.Download(uint(boardUid), uint(fileUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	// 일회용 토큰 발급 (5분 동안 접근 가능)
	token := uuid.New().String()
	expiry := time.Now().Add(1 * time.Minute)
	h.downloadTokenStorage[token] = DownloadToken{
		Name:   result.Name,
		Path:   result.Path,
		Expiry: expiry,
	}
	result.Path = fmt.Sprintf("/board/transfer?token=%s", token)
	return utils.Ok(c, result)
}

// 특정 사용자의 최근 활동(글, 댓글)들 가져오기
func (h *NuboBoardHandler) LatestUserContentHandler(c fiber.Ctx) error {
	uid, err := strconv.ParseUint(c.FormValue("targetUserUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	limit, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Board.GetLatestUserContents(uint(uid), uint(limit))
	return utils.Ok(c, result)
}

// 게시글 좋아하기 핸들러
func (h *NuboBoardHandler) LikePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	if actionUserUid < 1 {
		return utils.Err(c, "unauthorized error, only logged in user can like/unlike post", models.CODE_INVALID_TOKEN)
	}

	param := models.BoardViewLikeParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	param.UserUid = uint(actionUserUid)
	h.service.Board.LikeThisPost(param)
	return utils.Ok(c, nil)
}

// 게시글 이동 대상 목록 가져오는 핸들러
func (h *NuboBoardHandler) ListForMoveHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	boards, err := h.service.Board.GetBoardList(uint(boardUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, boards)
}

// 게시글 이동하기 핸들러
func (h *NuboBoardHandler) MovePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	targetBoardUid, err := strconv.ParseUint(c.FormValue("targetBoardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid target board uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	h.service.Board.MovePost(models.BoardMovePostParam{
		BoardViewCommonParam: models.BoardViewCommonParam{
			BoardUid: uint(boardUid),
			PostUid:  uint(postUid),
			UserUid:  uint(actionUserUid),
		},
		TargetBoardUid: uint(targetBoardUid),
	})
	return utils.Ok(c, nil)
}

// 게시글 삭제하기 핸들러
func (h *NuboBoardHandler) RemovePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	h.service.Board.RemovePost(uint(boardUid), uint(postUid), uint(actionUserUid))
	return utils.Ok(c, nil)
}

// (내부용) 다운로드용 토큰 정리하기
func (h *NuboBoardHandler) cleanupOldTokens() {
	now := time.Now()
	for oldToken, tokenData := range h.downloadTokenStorage {
		if now.After(tokenData.Expiry) {
			delete(h.downloadTokenStorage, oldToken)
		}
	}
}

// 일회용 토큰 값으로 파일 다운로드 하기
func (h *NuboBoardHandler) TransferHandler(c fiber.Ctx) error {
	h.cleanupOldTokens()

	token := c.Query("token")
	data, exists := h.downloadTokenStorage[token]

	if !exists {
		return c.Status(fiber.StatusForbidden).SendString("invalid token for downloading a file")
	}
	if time.Now().After(data.Expiry) {
		return c.Status(fiber.StatusForbidden).SendString("already expired token")
	}

	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	filePath := fmt.Sprintf(".%s", data.Path)
	return c.Download(filePath, data.Name)
}

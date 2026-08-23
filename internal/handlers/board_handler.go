package handlers

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
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
	OriginalImageHandler(c fiber.Ctx) error
	OriginalImageTransferHandler(c fiber.Ctx) error
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

type OriginalImageToken struct {
	Path   string
	Expiry time.Time
}

type NuboBoardHandler struct {
	service              *services.Service
	downloadTokenMu      sync.Mutex
	downloadTokenStorage map[string]DownloadToken
	originalImageTokenMu sync.Mutex
	originalImageTokens  map[string]OriginalImageToken
}

// services.Service 주입 받기
func NewNuboBoardHandler(service *services.Service) *NuboBoardHandler {
	return &NuboBoardHandler{
		service:              service,
		downloadTokenStorage: make(map[string]DownloadToken),
		originalImageTokens:  make(map[string]OriginalImageToken),
	}
}

// 게시글 목록 가져오기 핸들러
func (h *NuboBoardHandler) BoardListHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	if actionUserUid < 0 {
		actionUserUid = 0
	}
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
	if config.Type == models.BOARD_TRADE {
		result, err := h.service.Trade.GetList(parameter)
		if err != nil {
			return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
		}
		return utils.Ok(c, result)
	}

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
	if actionUserUid < 0 {
		actionUserUid = 0
	}

	param := models.BoardViewParam{}
	if err := c.Bind().Query(&param); err != nil {
		return utils.Err(c, "Invalid parameters", models.CODE_INVALID_PARAMETER)
	}
	param.UserUid = uint(actionUserUid)

	boardUid := h.service.Board.GetBoardUid(param.Id)
	if boardUid < 1 {
		return utils.Err(c, "Invalid board id, cannot find a board", models.CODE_INVALID_PARAMETER)
	}
	param.BoardUid = boardUid
	if h.service.Board.GetBoardConfig(boardUid).Type == models.BOARD_TRADE {
		result, err := h.service.Trade.GetView(param)
		if err != nil {
			return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
		}
		return utils.Ok(c, result)
	}

	result, err := h.service.Board.GetViewItem(param)
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
	h.storeDownloadToken(token, DownloadToken{
		Name:   result.Name,
		Path:   result.Path,
		Expiry: expiry,
	})
	result.Path = fmt.Sprintf("/board/transfer?token=%s", token)
	return utils.Ok(c, result)
}

// 게시물 보기 권한을 확인한 뒤 실제 저장 경로 대신 짧은 수명의 원본 스트리밍 URL을 발급한다.
func (h *NuboBoardHandler) OriginalImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	if actionUserUid < 0 {
		actionUserUid = 0
	}
	if actionUserUid > 0 && !h.service.Auth.CanAuthenticate(uint(actionUserUid)) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	boardUid, err := strconv.ParseUint(c.Query("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	fileUid, err := strconv.ParseUint(c.Query("fileUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid file uid, not a valid number", models.CODE_INVALID_PARAMETER)
	}
	result, err := h.service.Board.GetOriginalImage(uint(boardUid), uint(fileUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	token := uuid.New().String()
	h.storeOriginalImageToken(token, OriginalImageToken{
		Path:   result.Path,
		Expiry: time.Now().Add(2 * time.Minute),
	})
	result.Path = fmt.Sprintf("/board/original/transfer?token=%s", token)
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
	param := models.BoardViewLikeParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	param.UserUid = uint(actionUserUid)
	if err := h.service.Board.LikeThisPost(param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
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

	if err := h.service.Board.MovePost(models.BoardMovePostParam{
		BoardViewCommonParam: models.BoardViewCommonParam{
			BoardUid: uint(boardUid),
			PostUid:  uint(postUid),
			UserUid:  uint(actionUserUid),
		},
		TargetBoardUid: uint(targetBoardUid),
	}); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시글 삭제하기 핸들러
func (h *NuboBoardHandler) RemovePostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	param := models.RemovePostParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, "invalid parameters", models.CODE_INVALID_PARAMETER)
	}

	if err := h.service.Board.RemovePost(uint(param.BoardUid), uint(param.PostUid), uint(actionUserUid)); err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// (내부용) 다운로드용 토큰 정리하기
func (h *NuboBoardHandler) cleanupOldTokens() {
	h.downloadTokenMu.Lock()
	defer h.downloadTokenMu.Unlock()
	h.cleanupOldTokensLocked(time.Now())
}

func (h *NuboBoardHandler) cleanupOldTokensLocked(now time.Time) {
	for oldToken, tokenData := range h.downloadTokenStorage {
		if !now.Before(tokenData.Expiry) {
			delete(h.downloadTokenStorage, oldToken)
		}
	}
}

func (h *NuboBoardHandler) storeDownloadToken(token string, data DownloadToken) {
	h.downloadTokenMu.Lock()
	defer h.downloadTokenMu.Unlock()
	h.cleanupOldTokensLocked(time.Now())
	h.downloadTokenStorage[token] = data
}

func (h *NuboBoardHandler) consumeDownloadToken(token string, now time.Time) (DownloadToken, bool) {
	h.downloadTokenMu.Lock()
	defer h.downloadTokenMu.Unlock()
	h.cleanupOldTokensLocked(now)
	data, exists := h.downloadTokenStorage[token]
	if !exists {
		return DownloadToken{}, false
	}
	delete(h.downloadTokenStorage, token)
	return data, true
}

func (h *NuboBoardHandler) storeOriginalImageToken(token string, data OriginalImageToken) {
	h.originalImageTokenMu.Lock()
	defer h.originalImageTokenMu.Unlock()
	h.cleanupOriginalImageTokensLocked(time.Now())
	h.originalImageTokens[token] = data
}

func (h *NuboBoardHandler) originalImageToken(token string, now time.Time) (OriginalImageToken, bool) {
	h.originalImageTokenMu.Lock()
	defer h.originalImageTokenMu.Unlock()
	h.cleanupOriginalImageTokensLocked(now)
	data, exists := h.originalImageTokens[token]
	return data, exists
}

func (h *NuboBoardHandler) cleanupOriginalImageTokensLocked(now time.Time) {
	for token, data := range h.originalImageTokens {
		if !now.Before(data.Expiry) {
			delete(h.originalImageTokens, token)
		}
	}
}

// 원본은 브라우저 안에서 표시하며 byte range 요청을 허용한다.
// 토큰은 만료 전까지 재사용해 브라우저의 추가 range 요청도 처리한다.
func (h *NuboBoardHandler) OriginalImageTransferHandler(c fiber.Ctx) error {
	data, exists := h.originalImageToken(c.Query("token"), time.Now())
	if !exists {
		return c.Status(fiber.StatusForbidden).SendString("invalid token for viewing an image")
	}
	filePath, err := utils.UploadFilePath(data.Path)
	if err != nil || !utils.IsImage(filePath) {
		return c.Status(fiber.StatusNotFound).SendString("image not found")
	}
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	c.Set(fiber.HeaderContentDisposition, "inline")
	c.Set("X-Content-Type-Options", "nosniff")
	return c.SendFile(filePath, fiber.SendFile{ByteRange: true})
}

// 일회용 토큰 값으로 파일 다운로드 하기
func (h *NuboBoardHandler) TransferHandler(c fiber.Ctx) error {
	token := c.Query("token")
	data, exists := h.consumeDownloadToken(token, time.Now())

	if !exists {
		return c.Status(fiber.StatusForbidden).SendString("invalid token for downloading a file")
	}

	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	filePath := fmt.Sprintf(".%s", data.Path)
	return c.Download(filePath, data.Name)
}

package handlers

import (
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type EditorHandler interface {
	GetEditorConfigHandler(c fiber.Ctx) error
	LoadInsertImageHandler(c fiber.Ctx) error
	LoadPostHandler(c fiber.Ctx) error
	LoadThumbnailImageHandler(c fiber.Ctx) error
	ModifyPostHandler(c fiber.Ctx) error
	RemoveInsertImageHandler(c fiber.Ctx) error
	RemoveAttachedFileHandler(c fiber.Ctx) error
	SuggestionTitleHandler(c fiber.Ctx) error
	SuggestionHashtagHandler(c fiber.Ctx) error
	UploadInsertImageHandler(c fiber.Ctx) error
	WritePostHandler(c fiber.Ctx) error
}

type NuboEditorHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboEditorHandler(service *services.Service) *NuboEditorHandler {
	return &NuboEditorHandler{service: service}
}

// 에디터에서 게시판 설정, 카테고리 목록, 관리자 여부 가져오기
func (h *NuboEditorHandler) GetEditorConfigHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	result := h.service.Board.GetEditorConfig(boardUid, uint(actionUserUid))
	return utils.Ok(c, result)
}

// 게시글에 내가 삽입한 이미지들 불러오기 핸들러
func (h *NuboEditorHandler) LoadInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	lastUid, err := strconv.ParseUint(c.FormValue("lastUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	parameter := models.EditorInsertImageParam{
		BoardUid: uint(boardUid),
		LastUid:  uint(lastUid),
		UserUid:  uint(actionUserUid),
		Bunch:    uint(bunch),
	}
	result, err := h.service.Board.GetInsertedImages(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 글 수정을 위해 내가 작성한 게시글 정보 불러오기
func (h *NuboEditorHandler) LoadPostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	result, err := h.service.Board.LoadPost(uint(boardUid), uint(postUid), uint(actionUserUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, result)
}

// 첨부한 이미지 파일의 미리보기를 위한 썸네일 이미지 반환
func (h *NuboEditorHandler) LoadThumbnailImageHandler(c fiber.Ctx) error {
	fileUid, err := strconv.ParseUint(c.FormValue("fileUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	thumb, err := h.service.Board.GetThumbnailImage(uint(fileUid))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, thumb)
}

// 게시글 수정하기 핸들러
func (h *NuboEditorHandler) ModifyPostHandler(c fiber.Ctx) error {
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	parameter, err := utils.CheckWriteParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	err = h.service.Board.ModifyPost(models.EditorModifyParam{
		EditorWriteParam: parameter,
		PostUid:          uint(postUid),
	})
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, nil)
}

// 게시글에 삽입한 이미지 삭제하기 핸들러
func (h *NuboEditorHandler) RemoveInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	imageUid, err := strconv.ParseUint(c.FormValue("imageUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Board.RemoveInsertedImage(uint(imageUid), uint(actionUserUid))
	return utils.Ok(c, nil)
}

// 기존에 첨부했던 파일을 글 수정에서 삭제하기
func (h *NuboEditorHandler) RemoveAttachedFileHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	fileUid, err := strconv.ParseUint(c.FormValue("fileUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	h.service.Board.RemoveAttachedFile(models.EditorRemoveAttachedParam{
		BoardUid: uint(boardUid),
		PostUid:  uint(postUid),
		FileUid:  uint(fileUid),
		UserUid:  uint(actionUserUid),
	})
	return utils.Ok(c, nil)
}

// 글제목 추천 목록 반환하는 핸들러
func (h *NuboEditorHandler) SuggestionTitleHandler(c fiber.Ctx) error {
	input, err := url.QueryUnescape(c.FormValue("title"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	titles := h.service.Board.GetSuggestionTitles(input, uint(bunch))
	return utils.Ok(c, titles)
}

// 해시태그 추천 목록 반환하는 핸들러
func (h *NuboEditorHandler) SuggestionHashtagHandler(c fiber.Ctx) error {
	input, err := url.QueryUnescape(c.FormValue("tag"))
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}

	suggestions := h.service.Board.GetSuggestionTags(input, uint(bunch))
	return utils.Ok(c, suggestions)
}

// 게시글 내용에 이미지 삽입하는 핸들러
func (h *NuboEditorHandler) UploadInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	fileSizeLimit, _ := strconv.ParseInt(configs.Env.FileSizeLimit, 10, 32)
	form, err := c.MultipartForm()
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	images := form.File["images[]"]
	if len(images) < 1 {
		return utils.Err(c, "No files uploaded", models.CODE_INVALID_PARAMETER)
	}

	var totalFileSize int64
	for _, fileHeader := range images {
		totalFileSize += fileHeader.Size
	}
	if totalFileSize > fileSizeLimit {
		return utils.Err(c, "Uploaded files exceed size limitation", models.CODE_EXCEED_SIZE)
	}

	uploadedImages, err := h.service.Board.UploadInsertImage(uint(boardUid), uint(actionUserUid), images)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, uploadedImages)
}

// 게시글 작성하기 핸들러
func (h *NuboEditorHandler) WritePostHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckWriteParams(c)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}

	postUid, err := h.service.Board.WritePost(parameter)
	if err != nil {
		return utils.Err(c, err.Error(), models.CODE_FAILED_OPERATION)
	}
	return utils.Ok(c, postUid)
}

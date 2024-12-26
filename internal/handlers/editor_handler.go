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
	ModifyPostHandler(c fiber.Ctx) error
	RemoveInsertImageHandler(c fiber.Ctx) error
	RemoveAttachedFileHandler(c fiber.Ctx) error
	SuggestionHashtagHandler(c fiber.Ctx) error
	UploadInsertImageHandler(c fiber.Ctx) error
	WritePostHandler(c fiber.Ctx) error
}

type TsboardEditorHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardEditorHandler(service *services.Service) *TsboardEditorHandler {
	return &TsboardEditorHandler{service: service}
}

// 에디터에서 게시판 설정, 카테고리 목록, 관리자 여부 가져오기
func (h *TsboardEditorHandler) GetEditorConfigHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	result := h.service.Board.GetEditorConfig(boardUid, actionUserUid)
	return utils.Ok(c, result)
}

// 게시글에 내가 삽입한 이미지들 불러오기 핸들러
func (h *TsboardEditorHandler) LoadInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	lastUid, err := strconv.ParseUint(c.FormValue("lastUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid last uid, not a valid number")
	}
	bunch, err := strconv.ParseUint(c.FormValue("bunch"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid bunch, not a valid number")
	}

	parameter := models.EditorInsertImageParameter{
		BoardUid: uint(boardUid),
		LastUid:  uint(lastUid),
		UserUid:  actionUserUid,
		Bunch:    uint(bunch),
	}
	result, err := h.service.Board.GetInsertedImages(parameter)
	if err != nil {
		return utils.Err(c, "Unable to load a list of inserted images")
	}
	return utils.Ok(c, result)
}

// 글 수정을 위해 내가 작성한 게시글 정보 불러오기
func (h *TsboardEditorHandler) LoadPostHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}

	result, err := h.service.Board.LoadPost(uint(boardUid), uint(postUid), actionUserUid)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, result)
}

// 게시글 수정하기 핸들러
func (h *TsboardEditorHandler) ModifyPostHandler(c fiber.Ctx) error {
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}
	parameter, err := utils.CheckWriteParameters(c)
	if err != nil {
		return utils.Err(c, err.Error())
	}

	err = h.service.Board.ModifyPost(models.EditorModifyParameter{
		EditorWriteParameter: parameter,
		PostUid:              uint(postUid),
	})
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, nil)
}

// 게시글에 삽입한 이미지 삭제하기 핸들러
func (h *TsboardEditorHandler) RemoveInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	imageUid, err := strconv.ParseUint(c.FormValue("imageUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid image uid, not a valid number")
	}

	h.service.Board.RemoveInsertedImage(uint(imageUid), actionUserUid)
	return utils.Ok(c, nil)
}

// 기존에 첨부했던 파일을 글 수정에서 삭제하기
func (h *TsboardEditorHandler) RemoveAttachedFileHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	postUid, err := strconv.ParseUint(c.FormValue("postUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid post uid, not a valid number")
	}
	fileUid, err := strconv.ParseUint(c.FormValue("fileUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid file uid, not a valid number")
	}

	h.service.Board.RemoveAttachedFile(models.EditorRemoveAttachedParameter{
		BoardUid: uint(boardUid),
		PostUid:  uint(postUid),
		FileUid:  uint(fileUid),
		UserUid:  actionUserUid,
	})
	return utils.Ok(c, nil)
}

// 해시태그 추천 목록 반환하는 핸들러
func (h *TsboardEditorHandler) SuggestionHashtagHandler(c fiber.Ctx) error {
	input, err := url.QueryUnescape(c.FormValue("tag"))
	if err != nil {
		return utils.Err(c, "Invalid tag name")
	}
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	suggestions := h.service.Board.GetSuggestionTags(input, uint(bunch))
	return utils.Ok(c, suggestions)
}

// 게시글 내용에 이미지 삽입하는 핸들러
func (h *TsboardEditorHandler) UploadInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid board uid, not a valid number")
	}
	fileSizeLimit, _ := strconv.ParseInt(configs.Env.FileSizeLimit, 10, 32)
	form, err := c.MultipartForm()
	if err != nil {
		return utils.Err(c, "Failed to parse form")
	}
	images := form.File["images"]
	if len(images) < 1 {
		return utils.Err(c, "No files uploaded")
	}

	var totalFileSize int64
	for _, fileHeader := range images {
		totalFileSize += fileHeader.Size
	}
	if totalFileSize > fileSizeLimit {
		return utils.Err(c, "Uploaded files exceed size limitation")
	}

	uploadedImages, err := h.service.Board.UploadInsertImage(uint(boardUid), actionUserUid, images)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, uploadedImages)
}

// 게시글 작성하기 핸들러
func (h *TsboardEditorHandler) WritePostHandler(c fiber.Ctx) error {
	parameter, err := utils.CheckWriteParameters(c)
	if err != nil {
		return utils.Err(c, err.Error())
	}

	postUid, err := h.service.Board.WritePost(parameter)
	if err != nil {
		return utils.Err(c, err.Error())
	}
	return utils.Ok(c, postUid)
}

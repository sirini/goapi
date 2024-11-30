package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardHandler interface {
	BoardListHandler(c fiber.Ctx) error
	BoardViewHandler(c fiber.Ctx) error
	DownloadHandler(c fiber.Ctx) error
	GetEditorConfigHandler(c fiber.Ctx) error
	GalleryListHandler(c fiber.Ctx) error
	GalleryLoadPhotoHandler(c fiber.Ctx) error
	LikePostHandler(c fiber.Ctx) error
	ListForMoveHandler(c fiber.Ctx) error
	LoadInsertImageHandler(c fiber.Ctx) error
	LoadPostHandler(c fiber.Ctx) error
	MovePostHandler(c fiber.Ctx) error
	ModifyPostHandler(c fiber.Ctx) error
	RemoveInsertImageHandler(c fiber.Ctx) error
	RemoveAttachedFileHandler(c fiber.Ctx) error
	RemovePostHandler(c fiber.Ctx) error
	SuggestionHashtagHandler(c fiber.Ctx) error
	UploadInsertImageHandler(c fiber.Ctx) error
	WritePostHandler(c fiber.Ctx) error
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

// 에디터에서 게시판 설정, 카테고리 목록, 관리자 여부 가져오기
func (h *TsboardBoardHandler) GetEditorConfigHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	id := c.FormValue("id")
	boardUid := h.service.Board.GetBoardUid(id)
	result := h.service.Board.GetEditorConfig(boardUid, actionUserUid)
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

// 게시글에 내가 삽입한 이미지들 불러오기 핸들러
func (h *TsboardBoardHandler) LoadInsertImageHandler(c fiber.Ctx) error {
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
func (h *TsboardBoardHandler) LoadPostHandler(c fiber.Ctx) error {
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

// 게시글 수정하기 핸들러
func (h *TsboardBoardHandler) ModifyPostHandler(c fiber.Ctx) error {
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
func (h *TsboardBoardHandler) RemoveInsertImageHandler(c fiber.Ctx) error {
	actionUserUid := utils.ExtractUserUid(c.Get("Authorization"))
	imageUid, err := strconv.ParseUint(c.FormValue("imageUid"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid image uid, not a valid number")
	}

	h.service.Board.RemoveInsertedImage(uint(imageUid), actionUserUid)
	return utils.Ok(c, nil)
}

// 기존에 첨부했던 파일을 글 수정에서 삭제하기
func (h *TsboardBoardHandler) RemoveAttachedFileHandler(c fiber.Ctx) error {
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

// 해시태그 추천 목록 반환하는 핸들러
func (h *TsboardBoardHandler) SuggestionHashtagHandler(c fiber.Ctx) error {
	input := c.FormValue("tag")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	suggestions := h.service.Board.GetSuggestionTags(input, uint(bunch))
	return utils.Ok(c, suggestions)
}

// 게시글 내용에 이미지 삽입하는 핸들러
func (h *TsboardBoardHandler) UploadInsertImageHandler(c fiber.Ctx) error {
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
func (h *TsboardBoardHandler) WritePostHandler(c fiber.Ctx) error {
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

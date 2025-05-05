package services

import (
	"fmt"
	"mime/multipart"
	"os"
	"sync"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardService interface {
	Download(boardUid uint, fileUid uint, userUid uint) (models.BoardViewDownloadResult, error)
	GetBoardUid(id string) uint
	GetMaxUid() uint
	GetBoardConfig(boardUid uint) models.BoardConfig
	GetBoardList(boardUid uint, userUid uint) ([]models.BoardItem, error)
	GetEditorConfig(boardUid uint, userUid uint) models.EditorConfigResult
	GetGalleryGridItem(param models.BoardListParameter) ([]models.GalleryGridItem, error)
	GetGalleryList(param models.BoardListParameter) models.GalleryListResult
	GetGalleryPhotos(boardUid uint, postUid uint, userUid uint) (models.GalleryPhotoResult, error)
	GetInsertedImages(param models.EditorInsertImageParameter) (models.EditorInsertImageResult, error)
	GetListItem(param models.BoardListParameter) (models.BoardListResult, error)
	GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error)
	GetSuggestionTags(input string, bunch uint) []models.EditorTagItem
	GetViewItem(param models.BoardViewParameter) (models.BoardViewResult, error)
	LikeThisPost(param models.BoardViewLikeParameter)
	LoadPost(boardUid uint, postUid uint, userUid uint) (models.EditorLoadPostResult, error)
	MovePost(param models.BoardMovePostParameter)
	ModifyPost(param models.EditorModifyParameter) error
	RemoveAttachedFile(param models.EditorRemoveAttachedParameter)
	RemoveInsertedImage(imageUid uint, userUid uint)
	RemovePost(boardUid uint, postUid uint, userUid uint)
	SaveAttachments(boardUid uint, postUid uint, files []*multipart.FileHeader)
	SaveTags(boardUid uint, postUid uint, tags []string)
	SaveThumbnail(fileUid uint, postUid uint, path string) models.BoardThumbnail
	UploadInsertImage(boardUid uint, userUid uint, images []*multipart.FileHeader) ([]string, error)
	WritePost(param models.EditorWriteParameter) (uint, error)
}

type TsboardBoardService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardBoardService(repos *repositories.Repository) *TsboardBoardService {
	return &TsboardBoardService{repos: repos}
}

// 다운로드에 필요한 정보 반환
func (s *TsboardBoardService) Download(boardUid uint, fileUid uint, userUid uint) (models.BoardViewDownloadResult, error) {
	var result models.BoardViewDownloadResult
	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_DOWNLOAD)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	result = s.repos.BoardView.GetDownloadInfo(fileUid)
	fileSize := utils.GetFileSize(result.Path)
	if fileSize < 1 {
		return result, fmt.Errorf("file not found")
	}

	s.repos.User.UpdateUserPoint(userUid, uint(userPt+needPt))
	s.repos.User.UpdatePointHistory(models.UpdatePointParameter{
		UserUid:  userUid,
		BoardUid: boardUid,
		Action:   models.POINT_ACTION_VIEW,
		Point:    needPt,
	})
	return result, nil
}

// 게시판 고유 번호 가져오기
func (s *TsboardBoardService) GetBoardUid(id string) uint {
	return s.repos.Board.GetBoardUidById(id)
}

// 게시글 최대 고유번호 반환
func (s *TsboardBoardService) GetMaxUid() uint {
	return s.repos.Board.GetMaxUid(models.TABLE_POST)
}

// 게시판 설정값 가져오기
func (s *TsboardBoardService) GetBoardConfig(boardUid uint) models.BoardConfig {
	return s.repos.Board.GetBoardConfig(boardUid)
}

// 게시글 이동할 대상 게시판 목록 가져오기
func (s *TsboardBoardService) GetBoardList(boardUid uint, userUid uint) ([]models.BoardItem, error) {
	if isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid); !isAdmin {
		return nil, fmt.Errorf("unauthorized access")
	}
	boards := s.repos.BoardView.GetAllBoards()
	return boards, nil
}

// 게시판 설정 및 카테고리, 관리자 여부 반환
func (s *TsboardBoardService) GetEditorConfig(boardUid uint, userUid uint) models.EditorConfigResult {
	return models.EditorConfigResult{
		Config:     s.repos.Board.GetBoardConfig(boardUid),
		IsAdmin:    s.repos.Auth.CheckPermissionByUid(userUid, boardUid),
		Categories: s.repos.Board.GetBoardCategories(boardUid),
	}
}

// 갤러리에 사진 목록들 가져오기
func (s *TsboardBoardService) GetGalleryGridItem(param models.BoardListParameter) ([]models.GalleryGridItem, error) {
	posts := make([]models.BoardListItem, 0)
	var err error
	items := make([]models.GalleryGridItem, 0)

	if len(param.Keyword) < 2 {
		posts, err = s.repos.Board.GetNormalPosts(param)
	} else {
		switch param.Option {
		case models.SEARCH_TAG:
			posts, err = s.repos.Board.FindPostsByHashtag(param)
		case models.SEARCH_CATEGORY:
		case models.SEARCH_WRITER:
			posts, err = s.repos.Board.FindPostsByNameCategory(param)
		default:
			posts, err = s.repos.Board.FindPostsByTitleContent(param)
		}
	}
	if err != nil {
		return items, err
	}

	for _, post := range posts {
		images, _ := s.repos.BoardView.GetAttachedImages(post.Uid)
		item := models.GalleryGridItem{
			Uid:     post.Uid,
			Like:    post.Like,
			Liked:   post.Liked,
			Writer:  post.Writer,
			Comment: post.Comment,
			Title:   post.Title,
			Images:  images,
		}
		items = append(items, item)
	}
	return items, nil
}

// 갤러리 리스트 반환하기
func (s *TsboardBoardService) GetGalleryList(param models.BoardListParameter) models.GalleryListResult {
	images, _ := s.GetGalleryGridItem(param)
	return models.GalleryListResult{
		TotalPostCount: s.repos.Board.GetTotalPostCount(param.BoardUid),
		Config:         s.repos.Board.GetBoardConfig(param.BoardUid),
		Images:         images,
	}
}

// 게시글 번호에 해당하는 첨부 사진들 가져오기 (GetPost() 후 호출됨)
func (s *TsboardBoardService) GetGalleryPhotos(boardUid uint, postUid uint, userUid uint) (models.GalleryPhotoResult, error) {
	result := models.GalleryPhotoResult{}
	if isBanned := s.repos.BoardView.CheckBannedByWriter(postUid, userUid); isBanned {
		return result, fmt.Errorf("you have been blocked by writer")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_VIEW)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	images, err := s.repos.BoardView.GetAttachedImages(postUid)
	if err != nil {
		return result, err
	}
	result = models.GalleryPhotoResult{
		Config: s.repos.Board.GetBoardConfig(boardUid),
		Images: images,
	}
	return result, nil
}

// 게시글에 내가 삽입한 이미지 목록들 가져오기
func (s *TsboardBoardService) GetInsertedImages(param models.EditorInsertImageParameter) (models.EditorInsertImageResult, error) {
	result := models.EditorInsertImageResult{}
	userLv, _ := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, _ := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}

	images, err := s.repos.BoardEdit.GetInsertedImages(param)
	if err != nil {
		return result, err
	}

	result = models.EditorInsertImageResult{
		Images:          images,
		MaxImageUid:     s.repos.BoardEdit.GetMaxImageUid(param.BoardUid, param.UserUid),
		TotalImageCount: s.repos.BoardEdit.GetTotalImageCount(param.BoardUid, param.UserUid),
	}
	return result, nil
}

// 게시판 목록글들 가져오기
func (s *TsboardBoardService) GetListItem(param models.BoardListParameter) (models.BoardListResult, error) {
	items := make([]models.BoardListItem, 0)
	var err error

	notices, err := s.repos.Board.GetNoticePosts(param.BoardUid, param.UserUid)
	if err != nil {
		return models.BoardListResult{}, err
	}
	result := models.BoardListResult{}
	param.NoticeCount = uint(len(notices))

	if len(param.Keyword) < 2 {
		items, err = s.repos.Board.GetNormalPosts(param)
	} else {
		switch param.Option {
		case models.SEARCH_TAG:
			items, err = s.repos.Board.FindPostsByHashtag(param)
		case models.SEARCH_CATEGORY:
		case models.SEARCH_WRITER:
			items, err = s.repos.Board.FindPostsByNameCategory(param)
		case models.SEARCH_IMAGE_DESC:
			items, err = s.repos.Board.FindPostsByImageDescription(param)
		default:
			items, err = s.repos.Board.FindPostsByTitleContent(param)
		}
	}
	if err != nil {
		return result, err
	}

	result = models.BoardListResult{
		TotalPostCount: s.repos.Board.GetTotalPostCount(param.BoardUid),
		Config:         s.repos.Board.GetBoardConfig(param.BoardUid),
		Notices:        notices,
		Posts:          items,
		BlackList:      s.repos.User.GetUserBlackList(param.UserUid),
		IsAdmin:        s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid),
	}
	return result, nil
}

// 최근 사용된 해시태그 가져오기
func (s *TsboardBoardService) GetRecentTags(boardUid uint, limit uint) ([]models.BoardTag, error) {
	return s.repos.Board.GetRecentTags(boardUid, limit)
}

// 추천할 태그 목록들 가져오기
func (s *TsboardBoardService) GetSuggestionTags(input string, bunch uint) []models.EditorTagItem {
	return s.repos.BoardEdit.GetSuggestionTags(input, bunch)
}

// 게시글 가져오기
func (s *TsboardBoardService) GetViewItem(param models.BoardViewParameter) (models.BoardViewResult, error) {
	result := models.BoardViewResult{}
	if isBanned := s.repos.BoardView.CheckBannedByWriter(param.PostUid, param.UserUid); isBanned {
		return result, fmt.Errorf("you have been blocked by writer")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_VIEW)
	if userLv < needLv {
		return result, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	s.repos.User.UpdateUserPoint(param.UserUid, uint(userPt+needPt))
	s.repos.User.UpdatePointHistory(models.UpdatePointParameter{
		UserUid:  param.UserUid,
		BoardUid: param.BoardUid,
		Action:   models.POINT_ACTION_VIEW,
		Point:    needPt,
	})

	post, err := s.repos.BoardView.GetPost(param.PostUid, param.UserUid)
	if err != nil {
		return result, err
	}

	config := s.repos.Board.GetBoardConfig(param.BoardUid)
	result.Config = config
	result.Post = post
	result.Files = make([]models.BoardAttachment, 0)
	result.Images = make([]models.BoardAttachedImage, 0)

	if config.Level.Download <= userLv {
		files, err := s.repos.BoardView.GetAttachments(param.PostUid)
		if err != nil {
			return result, err
		}
		result.Files = files
	}

	images, err := s.repos.BoardView.GetAttachedImages(param.PostUid)
	if err != nil {
		return result, err
	}
	result.Images = images

	if param.UpdateHit {
		s.repos.BoardView.UpdatePostHit(param.PostUid)
	}

	if post.Status == models.CONTENT_SECRET {
		isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
		isWriter := post.Writer.UserUid == param.UserUid

		if !isAdmin && !isWriter {
			result.Post.Content = ""
			result.Files = make([]models.BoardAttachment, 0)
			result.Images = make([]models.BoardAttachedImage, 0)
		}
	}

	result.Tags = s.repos.BoardView.GetTags(param.PostUid)
	result.PrevPostUid = s.repos.BoardView.GetPrevPostUid(param.BoardUid, param.PostUid)
	result.NextPostUid = s.repos.BoardView.GetNextPostUid(param.BoardUid, param.PostUid)
	result.WriterPosts, _ = s.repos.BoardView.GetWriterLatestPost(post.Writer.UserUid, param.Limit)
	result.WriterComments, _ = s.repos.BoardView.GetWriterLatestComment(post.Writer.UserUid, param.Limit)
	return result, nil
}

// 글 작성자에게 차단당했는지 확인
func (s *TsboardBoardService) IsBannedByWriter(postUid uint, viewerUid uint) bool {
	return s.repos.BoardView.CheckBannedByWriter(postUid, viewerUid)
}

// 게시글에 좋아요 클릭
func (s *TsboardBoardService) LikeThisPost(param models.BoardViewLikeParameter) {
	if isLiked := s.repos.BoardView.IsLikedPost(param.PostUid, param.UserUid); isLiked {
		s.repos.BoardView.UpdateLikePost(param)
	} else {
		s.repos.BoardView.InsertLikePost(param)
	}
}

// 게시글 수정 시 기존 정보들 가져오기
func (s *TsboardBoardService) LoadPost(boardUid uint, postUid uint, userUid uint) (models.EditorLoadPostResult, error) {
	result := models.EditorLoadPostResult{}
	post, err := s.repos.BoardView.GetPost(postUid, userUid)
	if err != nil {
		return result, err
	}
	isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
	if !isAdmin && !isAuthor {
		return result, fmt.Errorf("you have no permission to edit this post")
	}

	files, err := s.repos.BoardView.GetAttachments(postUid)
	if err != nil {
		return result, err
	}
	tags := s.repos.BoardView.GetTags(postUid)

	result.Post = post
	result.Files = files
	result.Tags = tags
	return result, nil
}

// 게시글 이동하기
func (s *TsboardBoardService) MovePost(param models.BoardMovePostParameter) {
	if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
		return
	}
	s.repos.BoardView.UpdatePostBoardUid(param.TargetBoardUid, param.PostUid)
}

// 게시글 수정하기
func (s *TsboardBoardService) ModifyPost(param models.EditorModifyParameter) error {
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return fmt.Errorf("only the author can edit this post")
	}

	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return fmt.Errorf("you have no permission to edit post")
	}

	if param.IsNotice {
		if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
			param.IsNotice = false
		}
	}
	s.repos.BoardView.RemovePostTags(param.PostUid)
	s.repos.BoardEdit.UpdatePost(param)
	s.SaveTags(param.BoardUid, param.PostUid, param.Tags)
	s.SaveAttachments(param.BoardUid, param.PostUid, param.Files)
	return nil
}

// 게시글 수정 시 첨부했던 파일 삭제하기
func (s *TsboardBoardService) RemoveAttachedFile(param models.EditorRemoveAttachedParameter) {
	isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, param.PostUid, param.UserUid)
	if !isAdmin && !isAuthor {
		return
	}

	filePath := s.repos.BoardEdit.FindAttachedPathByUid(param.FileUid)
	removes := s.repos.BoardView.RemoveAttachedFile(param.FileUid, filePath)

	for _, target := range removes {
		os.Remove("." + target)
	}
}

// 게시글에 삽입한 이미지 삭제하기
func (s *TsboardBoardService) RemoveInsertedImage(imageUid uint, userUid uint) {
	removePath := s.repos.BoardEdit.RemoveInsertedImage(imageUid, userUid)
	if len(removePath) > 0 {
		os.Remove("." + removePath)
	}
}

// 게시글 삭제하기
func (s *TsboardBoardService) RemovePost(boardUid uint, postUid uint, userUid uint) {
	isAdmin := s.repos.Auth.CheckPermissionByUid(userUid, boardUid)
	isAuthor := s.repos.BoardView.IsWriter(models.TABLE_POST, postUid, userUid)
	if !isAdmin && !isAuthor {
		return
	}

	s.repos.BoardView.RemovePost(postUid)
	s.repos.BoardView.RemoveComments(postUid)
	s.repos.BoardView.RemovePostTags(postUid)
	removes := s.repos.BoardView.RemoveAttachments(postUid)

	for _, path := range removes {
		_ = os.Remove("." + path)
	}
}

// 첨부파일들을 저장하기
func (s *TsboardBoardService) SaveAttachments(boardUid uint, postUid uint, files []*multipart.FileHeader) {
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)

		go func(f *multipart.FileHeader) {
			defer wg.Done()

			savedPath, err := utils.SaveAttachmentFile(f)
			if err != nil {
				return
			}
			fileUid := s.repos.BoardEdit.InsertFile(models.EditorSaveFileParameter{
				BoardUid: boardUid,
				PostUid:  postUid,
				Name:     utils.CutString(f.Filename, 100),
				Path:     savedPath[1:],
			})

			if utils.IsImage(f.Filename) {
				thumb, err := utils.SaveThumbnailImage(savedPath)
				if err != nil {
					return
				}
				s.repos.BoardEdit.InsertFileThumbnail(models.EditorSaveThumbnailParameter{
					BoardThumbnail: models.BoardThumbnail{
						Large: thumb.Large[1:],
						Small: thumb.Small[1:],
					},
					FileUid: fileUid,
					PostUid: postUid,
				})
				exif := utils.ExtractExif(savedPath)
				s.repos.BoardEdit.InsertExif(fileUid, postUid, exif)
				if imgDesc, err := utils.AskImageDescription(thumb.Small); err == nil {
					s.repos.BoardEdit.InsertImageDescription(fileUid, postUid, imgDesc)
				}
			}
		}(file)
	}
	wg.Wait()
}

// 해시태그들 저장하기
func (s *TsboardBoardService) SaveTags(boardUid uint, postUid uint, tags []string) {
	for _, tag := range tags {
		tidyTag := utils.Purify(tag)
		if len(tidyTag) < 2 {
			continue
		}

		hashtagUid := s.repos.BoardEdit.FindTagUidByName(tag)
		if hashtagUid > 0 {
			s.repos.BoardEdit.UpdateTag(hashtagUid)
		} else {
			hashtagUid = s.repos.BoardEdit.InsertTag(boardUid, postUid, tag)
		}
		s.repos.BoardEdit.InsertPostHashtag(boardUid, postUid, hashtagUid)
	}
}

// 썸네일 이미지 생성 및 저장하기
func (s *TsboardBoardService) SaveThumbnail(fileUid uint, postUid uint, path string) models.BoardThumbnail {
	thumb, err := utils.SaveThumbnailImage(path)
	if err != nil {
		return thumb
	}
	s.repos.BoardEdit.InsertFileThumbnail(models.EditorSaveThumbnailParameter{
		BoardThumbnail: models.BoardThumbnail{
			Small: thumb.Small[1:],
			Large: thumb.Large[1:],
		},
		FileUid: fileUid,
		PostUid: postUid,
	})
	return thumb
}

// 게시글에 삽입할 이미지 파일 업로드 처리하기
func (s *TsboardBoardService) UploadInsertImage(boardUid uint, userUid uint, images []*multipart.FileHeader) ([]string, error) {
	imagePaths := make([]string, 0)
	if hasPerm := s.repos.Auth.CheckPermissionForAction(userUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return imagePaths, fmt.Errorf("you have no permission to write a new post")
	}

	if hasPerm := s.repos.BoardEdit.CheckWriterForBlog(boardUid, userUid); !hasPerm {
		return imagePaths, fmt.Errorf("only blog owner can write a new post")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(userUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(boardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return imagePaths, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return imagePaths, fmt.Errorf("not enough point")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	tempPaths := make([]string, 0)
	errors := make([]error, 0)

	for _, header := range images {
		wg.Add(1)

		go func(h *multipart.FileHeader) {
			defer wg.Done()

			file, err := h.Open()
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			defer file.Close()

			tempPath, err := utils.SaveUploadedFile(file, h.Filename)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			imagePath, err := utils.SaveInsertImage(tempPath)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			tempPaths = append(tempPaths, tempPath)
			imagePaths = append(imagePaths, imagePath[1:])
			mu.Unlock()

		}(header)
	}

	wg.Wait()

	if len(errors) > 0 {
		for _, tempPath := range tempPaths {
			os.Remove(tempPath)
		}
		return imagePaths, errors[0]
	}

	s.repos.BoardEdit.InsertImagePaths(boardUid, userUid, imagePaths)

	for _, tempPath := range tempPaths {
		os.Remove(tempPath)
	}
	return imagePaths, nil
}

// 새 게시글 작성하기
func (s *TsboardBoardService) WritePost(param models.EditorWriteParameter) (uint, error) {
	if hasPerm := s.repos.Auth.CheckPermissionForAction(param.UserUid, models.USER_ACTION_WRITE_POST); !hasPerm {
		return models.FAILED, fmt.Errorf("you have no permission to write a new post")
	}
	if hasPerm := s.repos.BoardEdit.CheckWriterForBlog(param.BoardUid, param.UserUid); !hasPerm {
		return models.FAILED, fmt.Errorf("only blog owner can write a new post")
	}

	userLv, userPt := s.repos.User.GetUserLevelPoint(param.UserUid)
	needLv, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_WRITE)
	if userLv < needLv {
		return models.FAILED, fmt.Errorf("level restriction")
	}
	if needPt < 0 && userPt < utils.Abs(needPt) {
		return models.FAILED, fmt.Errorf("not enough point")
	}
	s.repos.User.UpdateUserPoint(param.UserUid, uint(userPt+needPt))

	if param.IsNotice {
		if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
			param.IsNotice = false
		}
	}

	postUid := s.repos.BoardEdit.InsertPost(param)
	s.SaveTags(param.BoardUid, postUid, param.Tags)
	s.SaveAttachments(param.BoardUid, postUid, param.Files)

	return postUid, nil
}

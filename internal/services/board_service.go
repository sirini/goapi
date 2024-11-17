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
	GetSuggestionTags(input string, bunch uint) []models.EditorTagItem
	GetViewItem(param models.BoardViewParameter) (models.BoardViewResult, error)
	LikeThisPost(param models.BoardViewLikeParameter)
	MovePost(param models.BoardMovePostParameter)
	RemoveInsertedImage(imageUid uint, userUid uint)
	RemovePost(boardUid uint, postUid uint, userUid uint)
	UploadInsertImage(boardUid uint, userUid uint, images []*multipart.FileHeader) ([]string, error)
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
	return s.repos.Board.GetMaxUid()
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
		Config:  s.repos.Board.GetBoardConfig(boardUid),
		IsAdmin: s.repos.Auth.CheckPermissionByUid(userUid, boardUid),
	}
}

// 갤러리에 사진 목록들 가져오기
func (s *TsboardBoardService) GetGalleryGridItem(param models.BoardListParameter) ([]models.GalleryGridItem, error) {
	var (
		posts []models.BoardListItem
		err   error
	)
	items := []models.GalleryGridItem{}

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
		return result, fmt.Errorf("unable to get attached images")
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
	var (
		items []models.BoardListItem
		err   error
	)

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
		Posts:          items,
		BlackList:      s.repos.User.GetUserBlackList(param.UserUid),
		IsAdmin:        s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid),
	}
	return result, nil
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
	if post.Status == models.CONTENT_SECRET {
		if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
			return result, fmt.Errorf("you don't have permission to open this post")
		}
	}

	config := s.repos.Board.GetBoardConfig(param.BoardUid)
	result.Config = config
	result.Post = post

	if config.Level.Download <= userLv {
		files, err := s.repos.BoardView.GetAttachments(param.PostUid)
		if err != nil {
			return result, fmt.Errorf("unable to get attachments")
		}
		result.Files = files
	}

	images, err := s.repos.BoardView.GetAttachedImages(param.PostUid)
	if err != nil {
		return result, fmt.Errorf("unable to get attached images")
	}
	result.Images = images

	if param.UpdateHit {
		s.repos.BoardView.UpdatePostHit(param.PostUid)
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

// 게시글 이동하기
func (s *TsboardBoardService) MovePost(param models.BoardMovePostParameter) {
	if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
		return
	}
	s.repos.BoardView.UpdatePostBoardUid(param.TargetBoardUid, param.PostUid)
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
	s.repos.BoardView.RemovePostTags(postUid)
	removes := s.repos.BoardView.RemoveAttachments(postUid)

	for _, path := range removes {
		_ = os.Remove("." + path)
	}
}

// 게시글에 삽입할 이미지 파일 업로드 처리하기
func (s *TsboardBoardService) UploadInsertImage(boardUid uint, userUid uint, images []*multipart.FileHeader) ([]string, error) {
	imagePaths := make([]string, 0)
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
		go func(header *multipart.FileHeader) {
			defer wg.Done()

			file, err := header.Open()
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			defer file.Close()

			tempPath, err := utils.SaveUploadedFile(file, header.Filename)
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

	s.repos.BoardEdit.InsertImagePath(boardUid, userUid, imagePaths)

	for _, tempPath := range tempPaths {
		os.Remove(tempPath)
	}
	return imagePaths, nil
}

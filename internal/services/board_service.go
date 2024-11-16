package services

import (
	"fmt"

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
	GetGalleryGridItem(param models.BoardListParameter) ([]models.GalleryGridItem, error)
	GetGalleryList(param models.BoardListParameter) models.GalleryListResult
	GetListItem(param models.BoardListParameter) (models.BoardListResult, error)
	GetViewItem(param models.BoardViewParameter) (models.BoardViewResult, error)
	LikeThisPost(param models.BoardViewLikeParameter)
	MovePost(param models.BoardMovePostParameter)
	RemovePost(boardUid uint, postUid uint, userUid uint)
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

// 게시글 가져오기
func (s *TsboardBoardService) GetViewItem(param models.BoardViewParameter) (models.BoardViewResult, error) {
	config := s.repos.Board.GetBoardConfig(param.BoardUid)
	level, point := s.repos.User.GetUserLevelPoint(param.UserUid)
	result := models.BoardViewResult{}

	if config.Level.View > level {
		return result, fmt.Errorf("level restriction, your level is %d but needs %d", level, config.Level.View)
	}

	if isBanned := s.repos.BoardView.CheckBannedByWriter(param.PostUid, param.UserUid); isBanned {
		return result, fmt.Errorf("you have been blocked by writer")
	}

	_, needPt := s.repos.BoardView.GetNeededLevelPoint(param.BoardUid, models.BOARD_ACTION_VIEW)
	if needPt < 0 && point < utils.Abs(needPt) {
		return result, fmt.Errorf("not enough point")
	}

	s.repos.User.UpdateUserPoint(param.UserUid, uint(point+needPt))
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

	result.Config = config
	result.Post = post

	if config.Level.Download <= level {
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
		utils.RemoveFile(path)
	}
}

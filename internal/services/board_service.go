package services

import (
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type BoardService interface {
	GetBoardUid(id string) uint
	GetMaxUid() uint
	GetBoardConfig(boardUid uint) *models.BoardConfig
	LoadListItem(param *models.BoardListParameter) (*models.BoardListResult, error)
	LoadViewItem(param *models.BoardViewParameter) (*models.BoardViewResult, error)
}

type TsboardBoardService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardBoardService(repos *repositories.Repository) *TsboardBoardService {
	return &TsboardBoardService{repos: repos}
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
func (s *TsboardBoardService) GetBoardConfig(boardUid uint) *models.BoardConfig {
	return s.repos.Board.GetBoardConfig(boardUid)
}

// 글 작성자에게 차단당했는지 확인
func (s *TsboardBoardService) IsBannedByWriter(postUid uint, viewerUid uint) bool {
	return s.repos.BoardView.CheckBannedByWriter(postUid, viewerUid)
}

// 게시판 목록글들 가져오기
func (s *TsboardBoardService) LoadListItem(param *models.BoardListParameter) (*models.BoardListResult, error) {
	var (
		items []*models.BoardListItem
		err   error
	)

	notices, err := s.repos.Board.GetNoticePosts(param.BoardUid, param.UserUid)
	if err != nil {
		return nil, err
	}
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
		return nil, err
	}

	result := &models.BoardListResult{
		TotalPostCount: s.repos.Board.GetTotalPostCount(param.BoardUid),
		Config:         s.repos.Board.GetBoardConfig(param.BoardUid),
		Posts:          items,
		BlackList:      s.repos.User.GetUserBlackList(param.UserUid),
		IsAdmin:        s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid),
	}
	return result, nil
}

// 게시글 가져오기
func (s *TsboardBoardService) LoadViewItem(param *models.BoardViewParameter) (*models.BoardViewResult, error) {
	config := s.repos.Board.GetBoardConfig(param.BoardUid)
	level, point := s.repos.User.GetUserLevelPoint(param.UserUid)

	if config.Level.View > level {
		return nil, fmt.Errorf("level restriction, your level is %d but needs %d", level, config.Level.View)
	}

	if isBanned := s.repos.BoardView.CheckBannedByWriter(param.PostUid, param.UserUid); isBanned {
		return nil, fmt.Errorf("you have been blocked by writer")
	}

	amountPoint := s.repos.BoardView.GetNeededPoint(param.BoardUid, models.BOARD_ACTION_VIEW)
	if amountPoint < 0 && point < utils.Abs(amountPoint) {
		return nil, fmt.Errorf("not enough point")
	}

	s.repos.User.UpdateUserPoint(param.UserUid, uint(point+amountPoint))
	s.repos.User.UpdatePointHistory(&models.UpdatePointParameter{
		UserUid:  param.UserUid,
		BoardUid: param.BoardUid,
		Action:   models.POINT_ACTION_VIEW,
		Point:    amountPoint,
	})

	post, err := s.repos.BoardView.GetPost(param.PostUid, param.UserUid)
	if err != nil {
		return nil, err
	}
	if post.Status == models.POST_SECRET {
		if isAdmin := s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid); !isAdmin {
			return nil, fmt.Errorf("you don't have permission to open this post")
		}
	}

	result := &models.BoardViewResult{}
	result.Config = config
	result.Post = post

	if config.Level.Download <= level {
		files, err := s.repos.BoardView.GetAttachments(param.PostUid)
		if err != nil {
			return nil, fmt.Errorf("unable to get attachments")
		}
		result.Files = files
	}

	images, err := s.repos.BoardView.GetAttachedImages(param.PostUid)
	if err != nil {
		return nil, fmt.Errorf("unable to get attached images")
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

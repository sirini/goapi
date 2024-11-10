package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type BoardService interface {
	GetBoardUid(id string) uint
	GetMaxUid() uint
	GetBoardConfig(boardUid uint) *models.BoardConfig
	LoadListItem(param *models.BoardListParameter) (*models.BoardListResult, error)
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
	return s.repos.Board.LoadBoardConfig(boardUid)
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
		Config:         s.repos.Board.LoadBoardConfig(param.BoardUid),
		Posts:          items,
		BlackList:      s.repos.User.GetUserBlackList(param.UserUid),
		IsAdmin:        s.repos.Auth.CheckPermissionByUid(param.UserUid, param.BoardUid),
	}
	return result, nil
}

package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type HomeService interface {
	AddVisitorLog(userUid uint)
	GetSidebarLinks() ([]models.HomeSidebarGroupResult, error)
	GetLatestPosts(param *models.HomePostParameter) ([]*models.BoardHomePostItem, error)
}

type TsboardHomeService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardHomeService(repos *repositories.Repository) *TsboardHomeService {
	return &TsboardHomeService{repos: repos}
}

// 방문자 접속 기록하기
func (s *TsboardHomeService) AddVisitorLog(userUid uint) {
	s.repos.Home.InsertVisitorLog(userUid)
}

// 사이드바 그룹/게시판들 목록 가져오기
func (s *TsboardHomeService) GetSidebarLinks() ([]models.HomeSidebarGroupResult, error) {
	return s.repos.Home.GetGroupBoardLinks()
}

// 지정된 게시글 번호 이하의 최근글들 가져오기
func (s *TsboardHomeService) GetLatestPosts(param *models.HomePostParameter) ([]*models.BoardHomePostItem, error) {
	var (
		posts []*models.HomePostItem
		err   error
	)

	if len(param.Keyword) < 2 {
		posts, err = s.repos.Home.GetLatestPosts(param)
	} else {
		switch param.Option {
		case models.SEARCH_TAG:
			posts, err = s.repos.Home.FindLatestPostsByTag(param)
		case models.SEARCH_CATEGORY:
		case models.SEARCH_WRITER:
			posts, err = s.repos.Home.FindLatestPostsByUserUidCatUid(param)
		default:
			posts, err = s.repos.Home.FindLatestPostsByTitleContent(param)
		}
	}
	if err != nil {
		return nil, err
	}

	var items []*models.BoardHomePostItem
	for _, post := range posts {
		settings := s.repos.Home.GetBoardBasicSettings(post.BoardUid)
		if len(settings.Id) < 2 {
			continue
		}

		item := &models.BoardHomePostItem{}
		item.Uid = post.Uid
		item.Title = post.Title
		item.Content = post.Content
		item.Submitted = post.Submitted
		item.Modified = post.Modified
		item.Hit = post.Hit
		item.Status = post.Status

		item.Id = settings.Id
		item.Type = settings.Type
		item.UseCategory = settings.UseCategory

		item.Category = s.repos.Board.GetCategoryByUid(post.CategoryUid)
		item.Cover = s.repos.Board.GetCoverImage(post.Uid)
		item.Comment = s.repos.Board.GetCountByTable(models.TABLE_COMMENT, post.Uid)
		item.Writer = s.repos.Board.GetWriterInfo(post.UserUid)
		item.Like = s.repos.Board.GetCountByTable(models.TABLE_POST_LIKE, post.Uid)
		item.Liked = s.repos.Board.CheckLikedPost(post.Uid, param.UserUid)

		items = append(items, item)
	}
	return items, nil
}

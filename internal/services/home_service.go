package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type HomeService interface {
	AddVisitorLog(userUid uint)
	GetLatestPosts(param models.HomePostParam) ([]models.BoardHomePostItem, error)
	GetSidebarLinks() ([]models.HomeSidebarGroupResult, error)
}

type NuboHomeService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboHomeService(repos *repositories.Repository) *NuboHomeService {
	return &NuboHomeService{repos: repos}
}

// 방문자 접속 기록하기
func (s *NuboHomeService) AddVisitorLog(userUid uint) {
	s.repos.Home.InsertVisitorLog(userUid)
}

// 지정된 게시글 번호 이하의 최근글들 가져오기
func (s *NuboHomeService) GetLatestPosts(param models.HomePostParam) ([]models.BoardHomePostItem, error) {
	items := make([]models.BoardHomePostItem, 0)
	posts := make([]models.HomePostItem, 0)
	var err error

	if len(param.Keyword) < 2 {
		posts, err = s.repos.Home.GetLatestPosts(param)
	} else {
		switch param.Option {
		case models.SEARCH_TAG:
			posts, err = s.repos.Home.FindLatestPostsByTag(param)
		case models.SEARCH_CATEGORY:
		case models.SEARCH_WRITER:
			posts, err = s.repos.Home.FindLatestPostsByUserUidCatUid(param)
		case models.SEARCH_IMAGE_DESC:
			posts, err = s.repos.Home.FindLatestPostsByImageDescription(param)
		default:
			posts, err = s.repos.Home.FindLatestPostsByTitleContent(param)
		}
	}
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		settings := s.repos.Home.GetBoardBasicSettings(post.BoardUid)
		if len(settings.Id) < 2 {
			continue
		}

		item := models.BoardHomePostItem{}
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
		item.Comment = s.repos.Board.GetCommentCount(post.Uid)
		item.Writer = s.repos.Board.GetWriterInfo(post.UserUid)
		item.Like = s.repos.Board.GetLikeCount(post.Uid)
		item.Liked = s.repos.Board.CheckLikedPost(post.Uid, param.UserUid)
		items = append(items, item)
	}
	return items, nil
}

// 사이드바 그룹/게시판들 목록 가져오기
func (s *NuboHomeService) GetSidebarLinks() ([]models.HomeSidebarGroupResult, error) {
	return s.repos.Home.GetGroupBoardLinks()
}

package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type HomeService interface {
	AddVisitorLog(userUid uint)
	GetSidebarLinks() ([]*models.HomeSidebarGroupResult, error)
	GetLatestPosts(param *models.BoardPostParameter) ([]*models.BoardFinalPostItem, error)
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
func (s *TsboardHomeService) GetSidebarLinks() ([]*models.HomeSidebarGroupResult, error) {
	return s.repos.Home.LoadGroupBoardLinks()
}

// 지정된 게시글 번호 이하의 최근글들 가져오기
func (s *TsboardHomeService) GetLatestPosts(param *models.BoardPostParameter) ([]*models.BoardFinalPostItem, error) {
	var (
		posts []*models.BoardPostItem
		err   error
	)

	if len(param.Keyword) < 2 {
		posts, err = s.repos.Board.LoadLatestPosts(param)
	} else {
		switch param.Option {
		case models.SEARCH_TAG:
			posts, err = s.repos.Board.FindLatestPostsByTag(param)
		case models.SEARCH_CATEGORY:
		case models.SEARCH_WRITER:
			posts, err = s.repos.Board.FindLatestPostsByUserUidCatUid(param)
		default:
			posts, err = s.repos.Board.FindLatestPostsByTitleContent(param)
		}
	}
	if err != nil {
		return nil, err
	}

	var items []*models.BoardFinalPostItem
	for _, post := range posts {
		settings := s.repos.Board.GetBoardSettings(post.BoardUid)
		if len(settings.Id) < 2 {
			continue
		}

		item := &models.BoardFinalPostItem{}
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

		item.Category = s.repos.Board.GetCategoryNameByUid(post.CategoryUid)
		item.Cover = s.repos.File.GetCoverImage(post.Uid)
		item.Comment = s.repos.Board.GetCountByTable(models.TABLE_COMMENT, post.Uid)
		item.Writer = s.repos.Board.GetWriterInfo(post.UserUid)
		item.Like = s.repos.Board.GetCountByTable(models.TABLE_POST_LIKE, post.Uid)
		item.Liked = s.repos.Board.CheckLikedPost(post.Uid, param.UserUid)

		items = append(items, item)
	}
	return items, nil
}

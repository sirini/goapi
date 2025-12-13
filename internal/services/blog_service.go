package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type BlogService interface {
	GetLatestPosts(boardUid uint, bunch uint) ([]models.HomePostItem, error)
}

type NuboBlogService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboBlogService(repos *repositories.Repository) *NuboBlogService {
	return &NuboBlogService{repos: repos}
}

// 최근 게시글들 반환하기
func (s *NuboBlogService) GetLatestPosts(boardUid uint, bunch uint) ([]models.HomePostItem, error) {
	maxUid := s.repos.Board.GetMaxUid(models.TABLE_POST)
	return s.repos.Home.GetLatestPosts(models.HomePostParam{
		SinceUid: maxUid,
		Bunch:    bunch,
		Option:   models.SEARCH_NONE,
		Keyword:  "",
		UserUid:  0,
		BoardUid: boardUid,
	})
}

package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type HomeService interface {
	AddVisitorLog(userUid uint)
	GetSidebarLinks() ([]*models.HomeSidebarGroupResult, error)
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

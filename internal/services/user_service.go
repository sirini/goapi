package services

import (
	"log"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type UserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewUserService(repos *repositories.Repository) *UserService {
	return &UserService{repos: repos}
}

// 사용자의 공개 정보 조회
func (s *UserService) GetUserInfo(userUid uint) (*models.UserInfo, error) {
	info, err := s.repos.UserRepo.FindUserInfoByUid(userUid)
	if err != nil {
		log.Fatal("Failed to get an user info, given: ", userUid)
	}

	return info, err
}

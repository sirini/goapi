package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type UserService interface {
	GetUserInfo(userUid uint) (*models.UserInfo, error)
	ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool
}

type TsboardUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardUserService(repos *repositories.Repository) *TsboardUserService {
	return &TsboardUserService{repos: repos}
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfo, error) {
	return s.repos.UserRepo.FindUserInfoByUid(userUid)
}

// 사용자가 특정 유저를 신고하기
func (s *TsboardUserService) ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool {
	isAllowedAction := s.repos.UserRepo.CheckPermissionForAction(actorUid, models.SEND_REPORT)
	if !isAllowedAction {
		return false
	}
	if wantBlock {
		s.repos.UserRepo.InsertBlackList(actorUid, targetUid)
	}
	s.repos.UserRepo.InsertReportUser(actorUid, targetUid, report)
	return true
}

// 사용자 로그인 처리하기
func (s *TsboardUserService) Signin(id string, pw string) bool {
	// TODO
	return false
}

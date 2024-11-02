package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type UserService interface {
	GetUserInfo(userUid uint) (*models.UserInfoResult, error)
	ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool
	ChangePassword(verifyUid uint, userCode string, newPassword string) bool
}

type TsboardUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardUserService(repos *repositories.Repository) *TsboardUserService {
	return &TsboardUserService{repos: repos}
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfoResult, error) {
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

// 비밀번호 변경하기
func (s *TsboardUserService) ChangePassword(verifyUid uint, userCode string, newPassword string) bool {
	id, code := s.repos.UserRepo.FindIDCodeByVerifyUid(verifyUid)
	if id == "" || code == "" {
		return false
	}
	if code != userCode {
		return false
	}
	userUid := s.repos.UserRepo.FindUserUidById(id)
	if userUid < 1 {
		return false
	}

	s.repos.UserRepo.UpdatePassword(userUid, newPassword)
	return true
}

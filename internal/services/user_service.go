package services

import (
	"os"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type UserService interface {
	ChangePassword(verifyUid uint, userCode string, newPassword string) bool
	ChangeUserInfo(info *models.UpdateUserInfoParameter, oldInfo *models.UserInfoResult)
	GetUserInfo(userUid uint) (*models.UserInfoResult, error)
	ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool
}

type TsboardUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardUserService(repos *repositories.Repository) *TsboardUserService {
	return &TsboardUserService{repos: repos}
}

// 비밀번호 변경하기
func (s *TsboardUserService) ChangePassword(verifyUid uint, userCode string, newPassword string) bool {
	id, code := s.repos.AuthRepo.FindIDCodeByVerifyUid(verifyUid)
	if id == "" || code == "" {
		return false
	}
	if code != userCode {
		return false
	}
	userUid := s.repos.AuthRepo.FindUserUidById(id)
	if userUid < 1 {
		return false
	}

	s.repos.UserRepo.UpdatePassword(userUid, newPassword)
	return true
}

// 사용자 정보 업데이트하기
func (s *TsboardUserService) ChangeUserInfo(param *models.UpdateUserInfoParameter, oldInfo *models.UserInfoResult) {
	if len(param.Password) == 64 {
		s.repos.UserRepo.UpdatePassword(param.UserUid, param.Password)
	}
	s.repos.UserRepo.UpdateUserInfoString(param.UserUid, param.Name, param.Signature)

	if param.Profile != nil && param.ProfileHandler.Size > 0 {
		_ = os.Remove("." + oldInfo.Profile)
		imagePath := utils.SaveUploadedFile(models.PROFILE, param.Profile, param.ProfileHandler.Filename)
		profilePath := utils.SaveProfileImage("." + imagePath)

		if len(profilePath) > 0 {
			s.repos.UserRepo.UpdateUserProfile(param.UserUid, profilePath)
			_ = os.Remove("." + imagePath)
		}
	}
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfoResult, error) {
	return s.repos.AuthRepo.FindUserInfoByUid(userUid)
}

// 사용자가 특정 유저를 신고하기
func (s *TsboardUserService) ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool {
	isAllowedAction := s.repos.AuthRepo.CheckPermissionForAction(actorUid, models.SEND_REPORT)
	if !isAllowedAction {
		return false
	}
	if wantBlock {
		s.repos.UserRepo.InsertBlackList(actorUid, targetUid)
	}
	s.repos.UserRepo.InsertReportUser(actorUid, targetUid, report)
	return true
}

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
	ChangeUserPermission(actionUserUid uint, perm *models.UserPermissionReportResult)
	GetUserInfo(userUid uint) (*models.UserInfoResult, error)
	GetUserPermission(userUid uint) *models.UserPermissionReportResult
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

// 사용자 정보 변경하기
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

// 사용자 권한 변경하기
func (s *TsboardUserService) ChangeUserPermission(actionUserUid uint, perm *models.UserPermissionReportResult) {
	targetUserUid := perm.UserUid
	permission := &perm.UserPermissionResult

	isPermAdded := s.repos.UserRepo.IsPermissionAdded(targetUserUid)
	if isPermAdded {
		s.repos.UserRepo.UpdateUserPermission(targetUserUid, permission)
	} else {
		s.repos.UserRepo.InsertUserPermission(targetUserUid, permission)
	}

	isReported := s.repos.UserRepo.IsUserReported(targetUserUid)
	if isReported {
		s.repos.UserRepo.UpdateReportResponse(targetUserUid, perm.Response)
	} else {
		s.repos.UserRepo.InsertReportResponse(actionUserUid, targetUserUid, perm.Response)
	}

	s.repos.UserRepo.UpdateUserBlocked(targetUserUid, !perm.Login)
	s.repos.UserRepo.InsertNewChat(actionUserUid, targetUserUid, perm.Response)
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfoResult, error) {
	return s.repos.AuthRepo.FindUserInfoByUid(userUid)
}

// 사용자의 권한 조회
func (s *TsboardUserService) GetUserPermission(userUid uint) *models.UserPermissionReportResult {
	var result models.UserPermissionReportResult
	permission := s.repos.UserRepo.LoadUserPermission(userUid)
	isBlocked := s.repos.UserRepo.IsBlocked(userUid)
	response := s.repos.UserRepo.GetReportResponse(userUid)

	result.WritePost = permission.WritePost
	result.WriteComment = permission.WriteComment
	result.SendChatMessage = permission.SendChatMessage
	result.SendReport = permission.SendReport
	result.Login = !isBlocked
	result.UserUid = userUid
	result.Response = response

	return &result
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

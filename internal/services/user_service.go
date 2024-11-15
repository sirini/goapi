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
	GetUserLevelPoint(userUid uint) (int, int)
	GetUserPermission(userUid uint) *models.UserPermissionReportResult
	ReportTargetUser(actionUserUid uint, targetUserUid uint, wantBlock bool, report string) bool
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
	id, code := s.repos.Auth.FindIDCodeByVerifyUid(verifyUid)
	if id == "" || code == "" {
		return false
	}
	if code != userCode {
		return false
	}
	userUid := s.repos.Auth.FindUserUidById(id)
	if userUid < 1 {
		return false
	}

	s.repos.User.UpdatePassword(userUid, newPassword)
	return true
}

// 사용자 정보 변경하기
func (s *TsboardUserService) ChangeUserInfo(param *models.UpdateUserInfoParameter, oldInfo *models.UserInfoResult) {
	if len(param.Password) == 64 {
		s.repos.User.UpdatePassword(param.UserUid, param.Password)
	}
	s.repos.User.UpdateUserInfoString(param.UserUid, utils.Escape(param.Name), utils.Escape(param.Signature))

	if param.Profile != nil && param.ProfileHandler.Size > 0 {
		_ = os.Remove("." + oldInfo.Profile)
		imagePath := utils.SaveUploadedFile(models.UPLOAD_PROFILE, param.Profile, param.ProfileHandler.Filename)
		profilePath := utils.SaveProfileImage("." + imagePath)

		if len(profilePath) > 0 {
			s.repos.User.UpdateUserProfile(param.UserUid, profilePath)
			_ = os.Remove("." + imagePath)
		}
	}
}

// 사용자 권한 변경하기
func (s *TsboardUserService) ChangeUserPermission(actionUserUid uint, perm *models.UserPermissionReportResult) {
	targetUserUid := perm.UserUid
	permission := &perm.UserPermissionResult

	isPermAdded := s.repos.User.IsPermissionAdded(targetUserUid)
	if isPermAdded {
		s.repos.User.UpdateUserPermission(targetUserUid, permission)
	} else {
		s.repos.User.InsertUserPermission(targetUserUid, permission)
	}

	isReported := s.repos.User.IsUserReported(targetUserUid)
	responseReport := utils.Escape(perm.Response)
	if isReported {
		s.repos.User.UpdateReportResponse(targetUserUid, responseReport)
	} else {
		s.repos.User.InsertReportResponse(actionUserUid, targetUserUid, responseReport)
	}

	s.repos.User.UpdateUserBlocked(targetUserUid, !perm.Login)
	s.repos.Chat.InsertNewChat(actionUserUid, targetUserUid, responseReport)
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfoResult, error) {
	return s.repos.Auth.FindUserInfoByUid(userUid)
}

// 사용자의 레벨과 보유 포인트 가져오기
func (s *TsboardUserService) GetUserLevelPoint(userUid uint) (int, int) {
	return s.repos.User.GetUserLevelPoint(userUid)
}

// 사용자의 권한 조회
func (s *TsboardUserService) GetUserPermission(userUid uint) *models.UserPermissionReportResult {
	var result models.UserPermissionReportResult
	permission := s.repos.User.LoadUserPermission(userUid)
	isBlocked := s.repos.User.IsBlocked(userUid)
	response := s.repos.User.GetReportResponse(userUid)

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
func (s *TsboardUserService) ReportTargetUser(actionUserUid uint, targetUserUid uint, wantBlock bool, report string) bool {
	isAllowedAction := s.repos.Auth.CheckPermissionForAction(actionUserUid, models.USER_ACTION_SEND_REPORT)
	if !isAllowedAction {
		return false
	}
	if wantBlock {
		s.repos.User.InsertBlackList(actionUserUid, targetUserUid)
	}
	s.repos.User.InsertReportUser(actionUserUid, targetUserUid, utils.Escape(report))
	return true
}

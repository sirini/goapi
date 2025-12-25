package services

import (
	"fmt"
	"os"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type UserService interface {
	ChangePassword(verifyUid uint, userCode string, newPassword string) bool
	ChangeUserInfo(info models.UpdateUserInfoParam) error
	ChangeUserPermission(actionUserUid uint, perm models.UserPermissionReportResult) error
	GetUserInfo(userUid uint) (models.UserInfoResult, error)
	GetUserLevelPoint(userUid uint) (int, int)
	GetUserPermission(actionUserUid uint, targetUserUid uint) models.UserPermissionReportResult
	ReportTargetUser(param models.UserReportParam) bool
}

type NuboUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboUserService(repos *repositories.Repository) *NuboUserService {
	return &NuboUserService{repos: repos}
}

// 비밀번호 변경하기
func (s *NuboUserService) ChangePassword(verifyUid uint, userCode string, newPassword string) bool {
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
func (s *NuboUserService) ChangeUserInfo(param models.UpdateUserInfoParam) error {
	if len(param.Password) == 64 {
		s.repos.User.UpdatePassword(param.UserUid, param.Password)
	}
	s.repos.User.UpdateUserInfoString(param.UserUid, utils.Escape(param.Name), utils.Escape(param.Signature))

	if param.Profile != nil {
		file, err := param.Profile.Open()
		if err == nil {
			defer file.Close()
		}

		if param.Profile.Size > 0 {
			tempPath, err := utils.SaveUploadedFile(file, param.Profile.Filename)
			if err != nil {
				return err
			}
			profilePath, err := utils.SaveProfileImage(tempPath)
			if err != nil {
				os.Remove(tempPath)
				return err
			}

			s.repos.User.UpdateUserProfile(param.UserUid, profilePath[1:])
			err = os.Remove("." + param.OldProfile)
			if err != nil {
				return err
			}
			err = os.Remove(tempPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 사용자 권한 변경하기
func (s *NuboUserService) ChangeUserPermission(actionUserUid uint, perm models.UserPermissionReportResult) error {
	if isAdmin := s.repos.Auth.CheckPermissionByUid(actionUserUid, 0); !isAdmin {
		return fmt.Errorf("unauthorized access")
	}
	targetUserUid := perm.UserUid
	permission := perm.UserPermissionResult

	isPermAdded := s.repos.User.IsPermissionAdded(targetUserUid)
	if isPermAdded {
		err := s.repos.User.UpdateUserPermission(targetUserUid, permission)
		if err != nil {
			return err
		}
	} else {
		err := s.repos.User.InsertUserPermission(targetUserUid, permission)
		if err != nil {
			return err
		}
	}

	isReported := s.repos.User.IsUserReported(targetUserUid)
	responseReport := utils.Escape(perm.Response)
	if isReported {
		err := s.repos.User.UpdateReportResponse(targetUserUid, responseReport)
		if err != nil {
			return err
		}
	} else {
		err := s.repos.User.InsertReportResponse(actionUserUid, targetUserUid, responseReport)
		if err != nil {
			return err
		}
	}

	err := s.repos.User.UpdateUserBlocked(targetUserUid, !perm.Login)
	if err != nil {
		return err
	}
	s.repos.Chat.InsertNewChat(actionUserUid, targetUserUid, responseReport)
	return nil
}

// 사용자의 공개 정보 조회
func (s *NuboUserService) GetUserInfo(userUid uint) (models.UserInfoResult, error) {
	return s.repos.Auth.FindUserInfoByUid(userUid)
}

// 사용자의 레벨과 보유 포인트 가져오기
func (s *NuboUserService) GetUserLevelPoint(userUid uint) (int, int) {
	return s.repos.User.GetUserLevelPoint(userUid)
}

// 사용자의 권한 조회
func (s *NuboUserService) GetUserPermission(actionUserUid uint, targetUserUid uint) models.UserPermissionReportResult {
	result := models.UserPermissionReportResult{}
	if isAdmin := s.repos.Auth.CheckPermissionByUid(actionUserUid, 0); !isAdmin {
		return result
	}

	permission := s.repos.User.LoadUserPermission(targetUserUid)
	isBlocked := s.repos.User.IsBlocked(targetUserUid)
	response := s.repos.User.GetReportResponse(targetUserUid)

	result.WritePost = permission.WritePost
	result.WriteComment = permission.WriteComment
	result.SendChatMessage = permission.SendChatMessage
	result.SendReport = permission.SendReport
	result.Login = !isBlocked
	result.UserUid = targetUserUid
	result.Response = response

	return result
}

// 사용자가 특정 유저를 신고하기
func (s *NuboUserService) ReportTargetUser(param models.UserReportParam) bool {
	isAllowedAction := s.repos.Auth.CheckPermissionForAction(param.ActionUserUid, models.USER_ACTION_SEND_REPORT)
	if !isAllowedAction {
		return false
	}
	if param.CheckedBlackList {
		s.repos.User.InsertBlackList(param.ActionUserUid, param.TargetUserUid)
	}
	s.repos.User.InsertReportUser(param)
	return true
}

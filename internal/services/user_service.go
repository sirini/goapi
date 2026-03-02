package services

import (
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CheckReportStatus(actionUserUid uint, targetUserUid uint) models.UserCheckReportResult
	ChangePassword(verifyUid uint, userCode string, newPassword string) bool
	ChangeUserInfo(info models.UpdateUserInfoParam) error
	ChangeUserPermission(actionUserUid uint, param models.UserPermissionManageParam) error
	ChangeUserProfile(userUid uint, profile *multipart.FileHeader, oldProfile string) error
	GetUserInfo(userUid uint) (models.UserInfoResult, error)
	GetUserLevelPoint(userUid uint) (int, int)
	GetUserPermission(actionUserUid uint, targetUserUid uint) models.UserPermissionManageParam
	ReportTargetUser(param models.UserReportParam) bool
	UpdateResponseToReport(actionUserUid uint, param models.UserPermissionManageParam) error
}

type NuboUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboUserService(repos *repositories.Repository) *NuboUserService {
	return &NuboUserService{repos: repos}
}

// 이미 신고한 사용자인지, 내 블랙리스트에 이미 들어가 있는지 확인
func (s *NuboUserService) CheckReportStatus(actionUserUid uint, targetUserUid uint) models.UserCheckReportResult {
	result := models.UserCheckReportResult{}
	result.IsReported = s.repos.User.IsReported(actionUserUid, targetUserUid)
	result.IsBannedByMe = s.repos.User.IsBannedByTarget(targetUserUid, actionUserUid)
	return result
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

	newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false
	}
	s.repos.Auth.UpdateUserPasswordHash(userUid, string(newBcryptHash))
	return true
}

// 사용자 정보 변경하기
func (s *NuboUserService) ChangeUserInfo(param models.UpdateUserInfoParam) error {
	s.repos.User.UpdateUserInfoString(param.UserUid, utils.Escape(param.Name), utils.Escape(param.Signature))
	if param.Profile != nil {
		return s.ChangeUserProfile(param.UserUid, param.Profile, param.OldProfile)
	}
	return nil
}

// 사용자 권한 변경하기
func (s *NuboUserService) ChangeUserPermission(actionUserUid uint, param models.UserPermissionManageParam) error {
	if isAdmin := s.repos.Auth.CheckPermissionByUid(actionUserUid, 0); !isAdmin {
		return fmt.Errorf("unauthorized access")
	}

	isPermAdded := s.repos.User.IsPermissionAdded(param.UserUid)
	if isPermAdded {
		if err := s.repos.User.UpdateUserPermission(param.UserUid, param.UserPermissionResult); err != nil {
			return err
		}
	} else {
		if err := s.repos.User.InsertUserPermission(param.UserUid, param.UserPermissionResult); err != nil {
			return err
		}
	}

	if err := s.repos.User.UpdateUserBlocked(param.UserUid, !param.Login); err != nil {
		return err
	}

	param.Response = strings.TrimSpace(param.Response)
	if len(param.Response) > 1 {
		return s.UpdateResponseToReport(actionUserUid, param)
	}
	return nil
}

// 사용자 프로필 이미지 변경하기
func (s *NuboUserService) ChangeUserProfile(userUid uint, profile *multipart.FileHeader, oldProfile string) error {
	if profile == nil {
		return fmt.Errorf("profile is empty")
	}

	file, err := profile.Open()
	if err == nil {
		defer file.Close()
	}

	if profile.Size > 0 {
		tempPath, err := utils.SaveUploadedFile(file, profile.Filename)
		if err != nil {
			return err
		}
		profilePath, err := utils.SaveProfileImage(tempPath)
		if err != nil {
			os.Remove(tempPath)
			return err
		}

		s.repos.User.UpdateUserProfile(userUid, profilePath[1:])
		if len(oldProfile) > 1 {
			if err := os.Remove("." + oldProfile); err != nil {
				return err
			}
		}
		if err := os.Remove(tempPath); err != nil {
			return err
		}
	}

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
func (s *NuboUserService) GetUserPermission(actionUserUid uint, targetUserUid uint) models.UserPermissionManageParam {
	result := models.UserPermissionManageParam{}
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

// 신고된 사용자에 대한 조치사항 업데이트
func (s *NuboUserService) UpdateResponseToReport(actionUserUid uint, param models.UserPermissionManageParam) error {
	isReported := s.repos.User.IsUserReported(param.UserUid)
	responseReport := utils.Escape(param.Response)
	if isReported {
		if err := s.repos.User.UpdateReportResponse(param.UserUid, responseReport); err != nil {
			return err
		}
	} else {
		if err := s.repos.User.InsertReportResponse(actionUserUid, param.UserUid, responseReport); err != nil {
			return err
		}
	}

	chatUid := s.repos.Chat.InsertNewChat(actionUserUid, param.UserUid, responseReport)
	if chatUid < 1 {
		return fmt.Errorf("failed to add a new chat message to let user know")
	}
	return nil
}

package services

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	AcknowledgeAchievements(userUid uint, badgeKeys []string) error
	CheckReportStatus(actionUserUid uint, targetUserUid uint) models.UserCheckReportResult
	BlockUser(actionUserUid uint, targetUserUid uint) error
	ChangePassword(verifyUid uint, userCode string, newPassword string) bool
	ChangeUserInfo(info models.UpdateUserInfoParam) error
	ChangeUserPermission(actionUserUid uint, param models.UserPermissionManageParam) error
	ChangeUserProfile(userUid uint, profile *multipart.FileHeader, oldProfile string) error
	DeleteAccount(userUid uint, confirmation string) error
	GetUserInfo(userUid uint) (models.UserInfoResult, error)
	GetUnannouncedAchievements(userUid uint) ([]models.UserBadge, error)
	GetUserLevelPoint(userUid uint) (int, int)
	GetUserPermission(actionUserUid uint, targetUserUid uint) models.UserPermissionManageParam
	ReportTargetUser(param models.UserReportParam) bool
	UnblockUser(actionUserUid uint, targetUserUid uint) error
	UpdateResponseToReport(actionUserUid uint, param models.UserPermissionManageParam) error
}

func (s *NuboUserService) GetUnannouncedAchievements(userUid uint) ([]models.UserBadge, error) {
	if userUid < 1 {
		return nil, fmt.Errorf("invalid user")
	}
	return s.repos.Badge.FindUnannouncedForUser(userUid, 10)
}

func (s *NuboUserService) AcknowledgeAchievements(userUid uint, badgeKeys []string) error {
	if userUid < 1 || len(badgeKeys) == 0 || len(badgeKeys) > 10 {
		return fmt.Errorf("invalid achievements")
	}
	seen := make(map[string]struct{}, len(badgeKeys))
	keys := make([]string, 0, len(badgeKeys))
	for _, key := range badgeKeys {
		key = strings.TrimSpace(key)
		if key == "" || len(key) > 80 {
			return fmt.Errorf("invalid achievement key")
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return s.repos.Badge.MarkAnnounced(userUid, keys, uint64(time.Now().UnixMilli()))
}

func (s *NuboUserService) BlockUser(actionUserUid uint, targetUserUid uint) error {
	if actionUserUid < 1 || targetUserUid < 1 || actionUserUid == targetUserUid {
		return fmt.Errorf("invalid block target")
	}
	return s.repos.User.InsertBlackList(actionUserUid, targetUserUid)
}

func (s *NuboUserService) UnblockUser(actionUserUid uint, targetUserUid uint) error {
	if actionUserUid < 1 || targetUserUid < 1 || actionUserUid == targetUserUid {
		return fmt.Errorf("invalid block target")
	}
	return s.repos.User.RemoveBlackList(actionUserUid, targetUserUid)
}

func (s *NuboUserService) DeleteAccount(userUid uint, confirmation string) error {
	if userUid < 1 || strings.TrimSpace(confirmation) != "DELETE" {
		return fmt.Errorf("invalid account deletion confirmation")
	}
	profile := s.repos.Auth.FindMyInfoByUid(userUid).Profile
	paths, err := s.repos.User.DeleteAccount(userUid)
	if err != nil {
		return err
	}
	for _, path := range paths {
		_ = utils.RemoveUploadFile(path)
	}
	if profile != "" {
		_ = utils.RemoveUploadFile(profile)
	}
	return nil
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
	newBcryptHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false
	}
	id, ok := s.repos.Auth.ConsumeVerificationCode(verifyUid, userCode, "")
	if !ok {
		return false
	}
	userUid := s.repos.Auth.FindUserUidById(id)
	if userUid < 1 {
		return false
	}
	return s.repos.Auth.UpdateUserPasswordHash(userUid, string(newBcryptHash)) == nil
}

// 사용자 정보 변경하기
func (s *NuboUserService) ChangeUserInfo(param models.UpdateUserInfoParam) error {
	if err := s.repos.User.UpdateUserInfoString(param.UserUid, utils.Escape(param.Name), utils.Escape(param.Signature)); err != nil {
		return err
	}
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
	if err != nil {
		return err
	}
	defer file.Close()

	if profile.Size > 0 {
		tempPath, err := utils.SaveUploadedFile(file, profile.Filename)
		if err != nil {
			return err
		}
		defer os.Remove(tempPath)
		profilePath, err := utils.SaveProfileImage(tempPath)
		if err != nil {
			return err
		}

		publicProfilePath, err := utils.PublicUploadPath(profilePath)
		if err != nil {
			_ = os.Remove(profilePath)
			return err
		}
		if err := s.repos.User.UpdateUserProfile(userUid, publicProfilePath); err != nil {
			_ = os.Remove(profilePath)
			return err
		}
		if len(oldProfile) > 1 {
			_ = utils.RemoveUploadFile(oldProfile)
		}
	}

	return nil
}

// 사용자의 공개 정보 조회
func (s *NuboUserService) GetUserInfo(userUid uint) (models.UserInfoResult, error) {
	info, err := s.repos.Auth.FindUserInfoByUid(userUid)
	if err != nil {
		return info, err
	}
	if s.repos.Badge == nil {
		return info, nil
	}
	badges, badgeErr := s.repos.Badge.FindForUser(userUid, false)
	if badgeErr != nil {
		log.Printf("badge: failed to load profile achievements for user %d: %v", userUid, badgeErr)
		return info, nil
	}
	info.Badges = badges
	return info, nil
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

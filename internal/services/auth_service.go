package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type AuthService interface {
	CheckEmailExists(id string) bool
	CheckNameExists(name string, userUid uint) bool
	CheckUserPermission(userUid uint, action models.UserAction) bool
	ChangeHashForPassword(userUid uint, newBcryptHash string)
	GetMyInfo(userUid uint) models.MyInfoResult
	GetUpdatedAccessToken(userUid uint, refreshToken string) (string, bool)
	GetUserAndHash(id string) (models.MyInfoResult, string)
	Logout(userUid uint)
	ResetPassword(id string, hostname string) bool
	Signin(id string, pw string) models.MyInfoResult
	Signup(param models.SignupParameter) (models.SignupResult, error)
	SaveTokensInCookie(c fiber.Ctx, userUid uint) error
	VerifyEmail(param models.VerifyParameter) bool
}

type TsboardAuthService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardAuthService(repos *repositories.Repository) *TsboardAuthService {
	return &TsboardAuthService{repos: repos}
}

// 이메일 중복 체크
func (s *TsboardAuthService) CheckEmailExists(id string) bool {
	return s.repos.User.IsEmailDuplicated(id)
}

// 이름 중복 체크
func (s *TsboardAuthService) CheckNameExists(name string, userUid uint) bool {
	return s.repos.User.IsNameDuplicated(name, userUid)
}

// 사용자 권한 확인하기
func (s *TsboardAuthService) CheckUserPermission(userUid uint, action models.UserAction) bool {
	return s.repos.Auth.CheckPermissionForAction(userUid, action)
}

// 사용자 비밀번호를 SHA256 해시값에서 Bcrypt 해시값으로 변경하기
func (s *TsboardAuthService) ChangeHashForPassword(userUid uint, newBcryptHash string) {
	s.repos.Auth.UpdateUserPasswordHash(userUid, newBcryptHash)
}

// 로그인 한 내 정보 가져오기
func (s *TsboardAuthService) GetMyInfo(userUid uint) models.MyInfoResult {
	return s.repos.Auth.FindMyInfoByUid(userUid)
}

// 리프레시 토큰이 유효할 경우 새로운 액세스 토큰 발급하기
func (s *TsboardAuthService) GetUpdatedAccessToken(userUid uint, refreshToken string) (string, bool) {
	if isValid := s.repos.Auth.CheckRefreshToken(userUid, refreshToken); !isValid {
		return "", false
	}

	accessHours, _ := configs.GetJWTAccessRefresh()
	newAccessToken, err := utils.GenerateAccessToken(userUid, accessHours)
	if err != nil {
		return "", false
	}
	return newAccessToken, true
}

// 사용자의 정보와 함께 기존에 저장된 비밀번호 해시값 가져오기
func (s *TsboardAuthService) GetUserAndHash(id string) (models.MyInfoResult, string) {
	userUid := s.repos.Auth.FindUserUidById(id)
	userInfo := s.repos.Auth.FindMyInfoByUid(userUid)
	storedHash := s.repos.Auth.FindUserPasswordByUid(userUid)
	return userInfo, storedHash
}

// 로그아웃하기
func (s *TsboardAuthService) Logout(userUid uint) {
	s.repos.Auth.ClearRefreshToken(userUid)
}

// 비밀번호 초기화하기
func (s *TsboardAuthService) ResetPassword(id string, hostname string) bool {
	userUid := s.repos.Auth.FindUserUidById(id)
	if userUid < 1 {
		return false
	}

	if configs.Env.GmailAppPassword == "" {
		message := strings.ReplaceAll(templates.ResetPasswordChat, "{{Id}}", id)
		message = strings.ReplaceAll(message, "{{Uid}}", strconv.Itoa(int(userUid)))
		insertId := s.repos.Chat.InsertNewChat(userUid, 1, message)
		if insertId < 1 {
			return false
		}
	} else {
		code := uuid.New().String()[:6]
		verifyUid := s.repos.Auth.SaveVerificationCode(id, code)
		body := strings.ReplaceAll(templates.ResetPasswordBody, "{{Host}}", hostname)
		body = strings.ReplaceAll(body, "{{Uid}}", strconv.Itoa(int(verifyUid)))
		body = strings.ReplaceAll(body, "{{Code}}", code)
		body = strings.ReplaceAll(body, "{{From}}", configs.Env.GmailID)
		title := strings.ReplaceAll(templates.ResetPasswordTitle, "{{Host}}", hostname)
		return utils.SendMail(id, title, body)
	}
	return true
}

// 사용자 로그인 처리하기
func (s *TsboardAuthService) Signin(id string, pw string) models.MyInfoResult {
	user := s.repos.Auth.FindMyInfoByIDPW(id, pw)
	if user.Uid < 1 {
		return user
	}

	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	accessToken, err := utils.GenerateAccessToken(user.Uid, accessHours)
	if err != nil {
		return user
	}

	refreshToken, err := utils.GenerateRefreshToken(refreshDays)
	if err != nil {
		return user
	}

	user.Token = accessToken
	user.Refresh = refreshToken
	s.repos.Auth.SaveRefreshToken(user.Uid, refreshToken)
	s.repos.Auth.UpdateUserSignin(user.Uid)
	return user
}

// 신규 회원 바로 가입 혹은 인증 메일 발송
func (s *TsboardAuthService) Signup(param models.SignupParameter) (models.SignupResult, error) {
	isDupId := s.repos.User.IsEmailDuplicated(param.ID)
	signupResult := models.SignupResult{}
	var target uint
	if isDupId {
		return signupResult, fmt.Errorf("email(%s) is already in use", param.ID)
	}

	name := utils.Escape(param.Name)
	isDupName := s.repos.User.IsNameDuplicated(name, 0)
	if isDupName {
		return signupResult, fmt.Errorf("name(%s) is already in use", name)
	}

	if configs.Env.GmailAppPassword == "" {
		target = s.repos.User.InsertNewUser(param.ID, param.Password, name)
		if target < 1 {
			return signupResult, fmt.Errorf("failed to add a new user")
		}
	} else {
		code := uuid.New().String()[:6]
		body := strings.ReplaceAll(templates.VerificationBody, "{{Host}}", param.Hostname)
		body = strings.ReplaceAll(body, "{{Name}}", name)
		body = strings.ReplaceAll(body, "{{Code}}", code)
		body = strings.ReplaceAll(body, "{{From}}", configs.Env.GmailID)
		subject := fmt.Sprintf("[%s] Your verification code: %s", param.Hostname, code)

		result := utils.SendMail(param.ID, subject, body)
		if result {
			target = s.repos.Auth.SaveVerificationCode(param.ID, code)
		}
	}

	signupResult = models.SignupResult{
		Sendmail: configs.Env.GmailAppPassword != "",
		Target:   target,
	}
	return signupResult, nil
}

// 로그인 성공 시 액세스 토큰과 리프레시 토큰들을 쿠키에 보관하기
func (s *TsboardAuthService) SaveTokensInCookie(c fiber.Ctx, userUid uint) error {
	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	authToken, err := utils.GenerateAccessToken(userUid, accessHours)
	if err != nil {
		return err
	}
	refreshToken, err := utils.GenerateRefreshToken(refreshDays)
	if err != nil {
		return err
	}

	utils.SaveCookie(c, "auth-token", authToken, 1)
	utils.SaveCookie(c, "auth-refresh", refreshToken, refreshDays)
	return nil
}

// 이메일 인증 완료하기
func (s *TsboardAuthService) VerifyEmail(param models.VerifyParameter) bool {
	result := s.repos.Auth.CheckVerificationCode(param)
	if result {
		s.repos.User.InsertNewUser(param.Id, param.Password, utils.Escape(param.Name))
		return true
	}
	return false
}

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
	"github.com/sirini/goapi/pkg/utils"
)

type AuthService interface {
	CheckEmailExists(id string) bool
	CheckNameExists(name string, userUid uint) bool
	CheckRefreshToken(userUid uint, refreshToken string) bool
	CheckUserPermission(userUid uint, action models.UserAction) bool
	ChangeHashForPassword(userUid uint, newBcryptHash string)
	GetMyInfo(userUid uint) models.MyInfoResult
	GetUserAndHash(id string) (models.MyInfoResult, string)
	Logout(userUid uint)
	ResetPassword(param models.ResetPasswordParam) bool
	Signin(id string, pw string) models.MyInfoResult
	Signup(param models.SignupParam) (models.SignupResult, error)
	SaveTokensInCookie(c fiber.Ctx, userUid uint) (string, string, error)
	VerifyEmail(param models.VerifyParam) bool
}

type NuboAuthService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewNuboAuthService(repos *repositories.Repository) *NuboAuthService {
	return &NuboAuthService{repos: repos}
}

// 이메일 중복 체크
func (s *NuboAuthService) CheckEmailExists(id string) bool {
	return s.repos.User.IsEmailDuplicated(id)
}

// 이름 중복 체크
func (s *NuboAuthService) CheckNameExists(name string, userUid uint) bool {
	return s.repos.User.IsNameDuplicated(name, userUid)
}

// 리프레시 토큰이 유효할 경우 새로운 액세스 토큰 발급하기
func (s *NuboAuthService) CheckRefreshToken(userUid uint, refreshToken string) bool {
	return s.repos.Auth.CheckRefreshToken(userUid, refreshToken)
}

// 사용자 권한 확인하기
func (s *NuboAuthService) CheckUserPermission(userUid uint, action models.UserAction) bool {
	return s.repos.Auth.CheckPermissionForAction(userUid, action)
}

// 사용자 비밀번호를 SHA256 해시값에서 Bcrypt 해시값으로 변경하기
func (s *NuboAuthService) ChangeHashForPassword(userUid uint, newBcryptHash string) {
	s.repos.Auth.UpdateUserPasswordHash(userUid, newBcryptHash)
}

// 로그인 한 내 정보 가져오기
func (s *NuboAuthService) GetMyInfo(userUid uint) models.MyInfoResult {
	return s.repos.Auth.FindMyInfoByUid(userUid)
}

// 사용자의 정보와 함께 기존에 저장된 비밀번호 해시값 가져오기
func (s *NuboAuthService) GetUserAndHash(id string) (models.MyInfoResult, string) {
	userUid := s.repos.Auth.FindUserUidById(id)
	userInfo := s.repos.Auth.FindMyInfoByUid(userUid)
	storedHash := s.repos.Auth.FindUserPasswordByUid(userUid)
	return userInfo, storedHash
}

// 로그아웃하기
func (s *NuboAuthService) Logout(userUid uint) {
	s.repos.Auth.ClearRefreshToken(userUid)
}

// 비밀번호 초기화하기
func (s *NuboAuthService) ResetPassword(param models.ResetPasswordParam) bool {
	userUid := s.repos.Auth.FindUserUidById(param.Email)
	if userUid < 1 {
		return false
	}

	code := uuid.New().String()[:6]
	verifyUid := s.repos.Auth.SaveVerificationCode(param.Email, code)
	body := strings.ReplaceAll(param.Template, "{{Code}}", code)
	body = strings.ReplaceAll(body, "{{UserUid}}", strconv.Itoa(int(verifyUid)))
	subject := fmt.Sprintf("[%s] Reset your password", param.Hostname)
	from := fmt.Sprintf("Admin <noreply@%s>", param.Hostname)
	isSent := utils.SendMail(param.Email, from, subject, body)

	if !isSent {
		chatTemplate := "Request to reset password from {{Id}} ({{Uid}})"
		message := strings.ReplaceAll(chatTemplate, "{{Id}}", param.Email)
		message = strings.ReplaceAll(message, "{{Uid}}", strconv.Itoa(int(userUid)))
		insertId := s.repos.Chat.InsertNewChat(userUid, 1, message)
		if insertId < 1 {
			return false
		}
	}
	return isSent
}

// 사용자 로그인 처리하기
func (s *NuboAuthService) Signin(id string, pw string) models.MyInfoResult {
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
func (s *NuboAuthService) Signup(param models.SignupParam) (models.SignupResult, error) {
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

	code := uuid.New().String()[:6]
	body := strings.ReplaceAll(param.Template, "{{Code}}", code)
	from := fmt.Sprintf("Admin <noreply@%s>", param.Hostname)
	subject := fmt.Sprintf("[%s] Verification code: %s", param.Hostname, code)
	isSent := utils.SendMail(param.ID, from, subject, body)

	// 메일 발송이 안되면 그냥 사용자 바로 추가 처리
	if !isSent {
		target = s.repos.User.InsertNewUser(param.ID, param.Password, name)
		if target < 1 {
			return signupResult, fmt.Errorf("failed to add a new user")
		}
	}

	signupResult = models.SignupResult{
		Sendmail: isSent,
		Target:   target,
	}
	return signupResult, nil
}

// 로그인 성공 시 액세스 토큰과 리프레시 토큰들을 쿠키에 보관하기
func (s *NuboAuthService) SaveTokensInCookie(c fiber.Ctx, userUid uint) (string, string, error) {
	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	authToken, err := utils.GenerateAccessToken(userUid, accessHours)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := utils.GenerateRefreshToken(refreshDays)
	if err != nil {
		return authToken, "", err
	}

	utils.SaveCookie(c, "nubo-auth-token", authToken, accessHours)
	utils.SaveCookie(c, "nubo-auth-refresh", refreshToken, refreshDays*24)
	return authToken, refreshToken, nil
}

// 이메일 인증 완료하기
func (s *NuboAuthService) VerifyEmail(param models.VerifyParam) bool {
	result := s.repos.Auth.CheckVerificationCode(param)
	if result {
		s.repos.User.InsertNewUser(param.ID, param.Password, utils.Escape(param.Name))
		return true
	}
	return false
}

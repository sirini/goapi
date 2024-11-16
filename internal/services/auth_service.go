package services

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type AuthService interface {
	CheckEmailExists(id string) bool
	CheckNameExists(name string) bool
	CheckUserPermission(userUid uint, action models.UserAction) bool
	GetMyInfo(userUid uint) models.MyInfoResult
	Logout(userUid uint)
	ResetPassword(id string, r *http.Request) bool
	Signin(id string, pw string) models.MyInfoResult
	Signup(id string, pw string, name string, r *http.Request) (models.SignupResult, error)
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
func (s *TsboardAuthService) CheckNameExists(name string) bool {
	return s.repos.User.IsNameDuplicated(name)
}

// 사용자 권한 확인하기
func (s *TsboardAuthService) CheckUserPermission(userUid uint, action models.UserAction) bool {
	return s.repos.Auth.CheckPermissionForAction(userUid, action)
}

// 로그인 한 내 정보 가져오기
func (s *TsboardAuthService) GetMyInfo(userUid uint) models.MyInfoResult {
	return s.repos.Auth.FindMyInfoByUid(userUid)
}

// 로그아웃하기
func (s *TsboardAuthService) Logout(userUid uint) {
	s.repos.Auth.ClearRefreshToken(userUid)
}

// 비밀번호 초기화하기
func (s *TsboardAuthService) ResetPassword(id string, r *http.Request) bool {
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
		body := strings.ReplaceAll(templates.ResetPasswordBody, "{{Host}}", r.Host)
		body = strings.ReplaceAll(body, "{{Uid}}", strconv.Itoa(int(verifyUid)))
		body = strings.ReplaceAll(body, "{{Code}}", code)
		body = strings.ReplaceAll(body, "{{From}}", configs.Env.GmailID)
		title := strings.ReplaceAll(templates.ResetPasswordTitle, "{{Host}}", r.Host)
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

	authToken, err := utils.GenerateAccessToken(user.Uid, 2)
	if err != nil {
		return user
	}

	refreshToken, err := utils.GenerateRefreshToken(1)
	if err != nil {
		return user
	}

	user.Token = authToken
	user.Refresh = refreshToken
	s.repos.Auth.SaveRefreshToken(user.Uid, refreshToken)
	s.repos.Auth.UpdateUserSignin(user.Uid)
	return user
}

// 신규 회원 바로 가입 혹은 인증 메일 발송
func (s *TsboardAuthService) Signup(id string, pw string, name string, r *http.Request) (models.SignupResult, error) {
	isDupId := s.repos.User.IsEmailDuplicated(id)
	signupResult := models.SignupResult{}
	var target uint
	if isDupId {
		return signupResult, fmt.Errorf("email(%s) is already in use", id)
	}

	name = utils.Escape(name)
	isDupName := s.repos.User.IsNameDuplicated(name)
	if isDupName {
		return signupResult, fmt.Errorf("name(%s) is already in use", name)
	}

	if configs.Env.GmailAppPassword == "" {
		target = s.repos.User.InsertNewUser(id, pw, name)
		if target < 1 {
			return signupResult, fmt.Errorf("failed to add a new user")
		}
	} else {
		code := uuid.New().String()[:6]
		body := strings.ReplaceAll(templates.VerificationBody, "{{Host}}", r.Host)
		body = strings.ReplaceAll(body, "{{Name}}", name)
		body = strings.ReplaceAll(body, "{{Code}}", code)
		body = strings.ReplaceAll(body, "{{From}}", configs.Env.GmailID)
		subject := fmt.Sprintf("[%s] Your verification code: %s", r.Host, code)

		result := utils.SendMail(id, subject, body)
		if result {
			target = s.repos.Auth.SaveVerificationCode(id, code)
		}
	}

	signupResult = models.SignupResult{
		Sendmail: configs.Env.GmailAppPassword != "",
		Target:   target,
	}
	return signupResult, nil
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

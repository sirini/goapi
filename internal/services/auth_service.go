package services

import (
	"fmt"
	"log"
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
	Signin(id string, pw string) *models.MyInfoResult
	Signup(id string, pw string, name string, r *http.Request) (*models.SignupResult, error)
	CheckEmailExists(id string) bool
	CheckNameExists(name string) bool
	VerifyEmail(param *models.VerifyParameter) bool
	ResetPassword(id string, r *http.Request) bool
}

type TsboardAuthService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardAuthService(repos *repositories.Repository) *TsboardAuthService {
	return &TsboardAuthService{repos: repos}
}

// 사용자 로그인 처리하기
func (s *TsboardAuthService) Signin(id string, pw string) *models.MyInfoResult {
	user := s.repos.UserRepo.FindMyInfoByIDPW(id, pw)
	if user.Uid < 1 {
		return &models.MyInfoResult{}
	}

	authToken, err := utils.GenerateAccessToken(user.Uid, 2)
	if err != nil {
		return &models.MyInfoResult{}
	}

	refreshToken, err := utils.GenerateRefreshToken(1)
	if err != nil {
		return &models.MyInfoResult{}
	}

	user.Token = authToken
	user.Refresh = refreshToken
	s.repos.UserRepo.SaveRefreshToken(user.Uid, refreshToken)
	s.repos.UserRepo.UpdateUserSignin(user.Uid)
	return user
}

// 신규 회원 바로 가입 혹은 인증 메일 발송
func (s *TsboardAuthService) Signup(id string, pw string, name string, r *http.Request) (*models.SignupResult, error) {
	isDupId := s.repos.UserRepo.IsEmailDuplicated(id)
	var target uint
	if isDupId {
		return &models.SignupResult{}, fmt.Errorf("Email(%s) is already in use", id)
	}

	isDupName := s.repos.UserRepo.IsNameDuplicated(name)
	if isDupName {
		return &models.SignupResult{}, fmt.Errorf("Name(%s) is already in use", name)
	}

	if configs.Env.GmailAppPassword == "" {
		target = s.repos.UserRepo.InsertNewUser(id, pw, name)
		if target < 1 {
			log.Fatalf("Failed to signup for %s (%s) : %s", id, name, pw)
			return &models.SignupResult{}, fmt.Errorf("Failed to add a new user")
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
			target = s.repos.UserRepo.SaveVerificationCode(id, code)
		}
	}
	return &models.SignupResult{
		Sendmail: configs.Env.GmailAppPassword != "",
		Target:   target,
	}, nil
}

// 이메일 중복 체크
func (s *TsboardAuthService) CheckEmailExists(id string) bool {
	return s.repos.UserRepo.IsEmailDuplicated(id)
}

// 이름 중복 체크
func (s *TsboardAuthService) CheckNameExists(name string) bool {
	return s.repos.UserRepo.IsNameDuplicated(name)
}

// 이메일 인증 완료하기
func (s *TsboardAuthService) VerifyEmail(param *models.VerifyParameter) bool {
	result := s.repos.UserRepo.CheckVerificationCode(param)
	if result {
		s.repos.UserRepo.InsertNewUser(param.Id, param.Password, param.Name)
		return true
	}
	return false
}

// 비밀번호 초기화하기
func (s *TsboardAuthService) ResetPassword(id string, r *http.Request) bool {
	userUid := s.repos.UserRepo.FindUserUidById(id)
	if userUid < 1 {
		return false
	}

	if configs.Env.GmailAppPassword == "" {
		message := strings.ReplaceAll(templates.ResetPasswordChat, "{{Id}}", id)
		message = strings.ReplaceAll(message, "{{Uid}}", strconv.Itoa(int(userUid)))
		insertId := s.repos.UserRepo.InsertNewChat(userUid, 1, message)
		if insertId < 1 {
			return false
		}
	} else {
		code := uuid.New().String()[:6]
		verifyUid := s.repos.UserRepo.SaveVerificationCode(id, code)
		body := strings.ReplaceAll(templates.ResetPasswordBody, "{{Host}}", r.Host)
		body = strings.ReplaceAll(body, "{{Uid}}", strconv.Itoa(int(verifyUid)))
		body = strings.ReplaceAll(body, "{{Code}}", code)
		body = strings.ReplaceAll(body, "{{From}}", configs.Env.GmailID)
		title := strings.ReplaceAll(templates.ResetPasswordTitle, "{{Host}}", r.Host)

		return utils.SendMail(id, title, body)
	}
	return true
}

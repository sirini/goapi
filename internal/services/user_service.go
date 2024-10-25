package services

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type UserService interface {
	GetUserInfo(userUid uint) (*models.UserInfoResult, error)
	ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool
	Signin(id string, pw string) *models.MyInfoResult
	Signup(id string, pw string, name string, r *http.Request) (*models.SignupResult, error)
	CheckEmailExists(id string) bool
	CheckNameExists(id string) bool
	VerifyEmail(param *models.VerifyParameter) bool
}

type TsboardUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardUserService(repos *repositories.Repository) *TsboardUserService {
	return &TsboardUserService{repos: repos}
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfoResult, error) {
	return s.repos.UserRepo.FindUserInfoByUid(userUid)
}

// 사용자가 특정 유저를 신고하기
func (s *TsboardUserService) ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool {
	isAllowedAction := s.repos.UserRepo.CheckPermissionForAction(actorUid, models.SEND_REPORT)
	if !isAllowedAction {
		return false
	}
	if wantBlock {
		s.repos.UserRepo.InsertBlackList(actorUid, targetUid)
	}
	s.repos.UserRepo.InsertReportUser(actorUid, targetUid, report)
	return true
}

// 사용자 로그인 처리하기
func (s *TsboardUserService) Signin(id string, pw string) *models.MyInfoResult {
	user := s.repos.UserRepo.FindMyInfoByIDPW(id, pw)
	if user.Uid < 1 {
		return &models.MyInfoResult{}
	}

	auth := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": user.Uid,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	authToken, err := auth.SignedString([]byte(configs.Env.JWTSecretKey))
	if err != nil {
		return &models.MyInfoResult{}
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().AddDate(0, 1, 0).Unix(),
	})
	refreshToken, err := refresh.SignedString([]byte(configs.Env.JWTSecretKey))
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
func (s *TsboardUserService) Signup(id string, pw string, name string, r *http.Request) (*models.SignupResult, error) {
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
func (s *TsboardUserService) CheckEmailExists(id string) bool {
	return s.repos.UserRepo.IsEmailDuplicated(id)
}

// 이름 중복 체크
func (s *TsboardUserService) CheckNameExists(name string) bool {
	return s.repos.UserRepo.IsNameDuplicated(name)
}

// 이메일 인증 완료하기
func (s *TsboardUserService) VerifyEmail(param *models.VerifyParameter) bool {
	result := s.repos.UserRepo.CheckVerificationCode(param)
	if result {
		s.repos.UserRepo.InsertNewUser(param.Id, param.Password, param.Name)
		return true
	}
	return false
}

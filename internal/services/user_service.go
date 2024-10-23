package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type UserService interface {
	GetUserInfo(userUid uint) (*models.UserInfo, error)
	ReportTargetUser(actorUid uint, targetUid uint, wantBlock bool, report string) bool
	Signin(id string, pw string) *models.MyInfo
}

type TsboardUserService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardUserService(repos *repositories.Repository) *TsboardUserService {
	return &TsboardUserService{repos: repos}
}

// 사용자의 공개 정보 조회
func (s *TsboardUserService) GetUserInfo(userUid uint) (*models.UserInfo, error) {
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
func (s *TsboardUserService) Signin(id string, pw string) *models.MyInfo {
	user := s.repos.UserRepo.FindMyInfoByIDPW(id, pw)
	if user.Uid < 1 {
		return &models.MyInfo{}
	}

	auth := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": user.Uid,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	authToken, err := auth.SignedString([]byte(configs.Env.JWTSecretKey))
	if err != nil {
		return &models.MyInfo{}
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().AddDate(0, 1, 0).Unix(),
	})
	refreshToken, err := refresh.SignedString([]byte(configs.Env.JWTSecretKey))
	if err != nil {
		return &models.MyInfo{}
	}

	user.Token = authToken
	user.Refresh = refreshToken
	s.repos.UserRepo.SaveRefreshToken(user.Uid, refreshToken)
	s.repos.UserRepo.UpdateUserSignin(user.Uid)
	return user
}

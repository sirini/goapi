package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type OAuthService interface {
	SaveProfileImage(userUid uint, profile string)
	RegisterOAuthUser(id string, name string, profile string) uint
	GenerateTokens(userUid uint) (string, string)
	SaveRefreshToken(userUid uint, token string)
	GetUserUid(id string) uint
	GetUserInfo(userUid uint) *models.MyInfoResult
}

type TsboardOAuthService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardOAuthService(repos *repositories.Repository) *TsboardOAuthService {
	return &TsboardOAuthService{repos: repos}
}

// OAuth 계정에 프로필 이미지가 있다면 가져와 저장하기
func (s *TsboardOAuthService) SaveProfileImage(userUid uint, profile string) {
	resp, err := http.Get(profile)
	if err != nil {
		log.Fatal("Unable to get an image from profile url")
		return
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to read data from body")
		return
	}

	imgOptions := bimg.Options{
		Width:   configs.Env.Number(configs.SIZE_PROFILE),
		Height:  configs.Env.Number(configs.SIZE_PROFILE),
		Quality: 100,
		Type:    bimg.AVIF,
	}

	newImg, err := bimg.NewImage(imgData).Process(imgOptions)
	if err != nil {
		log.Fatal("Failed to generate a new image")
		return
	}

	dirPath, err := utils.MakeSavePath("profile")
	if err != nil {
		log.Fatal("Unable to make directories")
		return
	}

	newSavePath := fmt.Sprintf("%s/%s.avif", dirPath, uuid.New().String())
	err = os.WriteFile(newSavePath, newImg, 0644)
	if err != nil {
		log.Fatal("Unable to write a image file")
		return
	}
}

// OAuth 로그인 시 미가입 상태이면 바로 등록해주기 (프로필도 있으면 함께)
func (s *TsboardOAuthService) RegisterOAuthUser(id string, name string, profile string) uint {
	pw := uuid.New().String()[:10]
	pw = utils.GetHashedString(pw)
	userUid := s.repos.UserRepo.InsertNewUser(id, pw, name)
	if userUid > 0 && profile != "" {
		s.SaveProfileImage(userUid, profile)
	}
	return userUid
}

// OAuth 로그인 후 액세스, 리프레시 토큰 생성해주기
func (s *TsboardOAuthService) GenerateTokens(userUid uint) (string, string) {
	auth, _ := utils.GenerateAccessToken(userUid, 2)
	refresh, _ := utils.GenerateRefreshToken(1)
	return auth, refresh
}

// 리프레시 토큰을 DB에 저장해주기
func (s *TsboardOAuthService) SaveRefreshToken(userUid uint, token string) {
	s.repos.UserRepo.SaveRefreshToken(userUid, token)
	s.repos.UserRepo.UpdateUserSignin(userUid)
}

// 회원 아이디(이메일)에 해당하는 고유 번호 반환
func (s *TsboardOAuthService) GetUserUid(id string) uint {
	return s.repos.UserRepo.FindUserUidById(id)
}

// 회원 정보 가져오기
func (s *TsboardOAuthService) GetUserInfo(userUid uint) *models.MyInfoResult {
	return s.repos.UserRepo.FindMyInfoByUid(userUid)
}

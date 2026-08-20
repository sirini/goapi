package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type mobileRefreshRepo struct {
	repositories.AuthRepository
	userUid uint
	old     string
	rotated bool
}

func (r *mobileRefreshRepo) FindMyInfoByUid(uid uint) models.MyInfoResult {
	if uid == r.userUid {
		return models.MyInfoResult{UserInfoResult: models.UserInfoResult{Uid: uid}}
	}
	return models.MyInfoResult{}
}

func (r *mobileRefreshRepo) RotateRefreshToken(userUid uint, oldRefreshToken, newRefreshToken string) bool {
	if userUid != r.userUid || oldRefreshToken != r.old || newRefreshToken == "" {
		return false
	}
	r.rotated = true
	return true
}

type legacySigninRepo struct {
	repositories.AuthRepository
	email        string
	legacyHash   string
	migratedUID  uint
	migratedHash string
}

func (r *legacySigninRepo) FindUserUidById(id string) uint {
	if id == r.email {
		return 7
	}
	return 0
}

func (r *legacySigninRepo) FindMyInfoByUid(uid uint) models.MyInfoResult {
	if uid == 7 {
		return models.MyInfoResult{UserInfoResult: models.UserInfoResult{Uid: uid}, Id: r.email}
	}
	return models.MyInfoResult{}
}

func (r *legacySigninRepo) FindUserPasswordByUid(uid uint) string {
	if uid == 7 {
		return r.legacyHash
	}
	return ""
}

func (r *legacySigninRepo) FindMyInfoByIDPW(id, password string) models.MyInfoResult {
	if id == r.email && password == r.legacyHash {
		return models.MyInfoResult{UserInfoResult: models.UserInfoResult{Uid: 7}, Id: id}
	}
	return models.MyInfoResult{}
}

func (*legacySigninRepo) SaveRefreshToken(uint, string) {}
func (*legacySigninRepo) UpdateUserSignin(uint)         {}
func (r *legacySigninRepo) UpdateUserPasswordHash(uid uint, password string) error {
	r.migratedUID = uid
	r.migratedHash = password
	return nil
}

func TestSigninMigratesLegacySHA256PasswordToBcrypt(t *testing.T) {
	previous := configs.Env
	configs.Env.JWTSecretKey = "test-secret"
	t.Cleanup(func() { configs.Env = previous })

	password := "Password!1"
	digest := sha256.Sum256([]byte(password))
	repo := &legacySigninRepo{
		email:      "member@example.com",
		legacyHash: hex.EncodeToString(digest[:]),
	}
	authService := services.NewNuboAuthService(&repositories.Repository{Auth: repo})
	handler := NewNuboAuthHandler(&services.Service{Auth: authService})
	app := fiber.New()
	app.Post("/signin", handler.SigninHandler)
	req := httptest.NewRequest("POST", "/signin", strings.NewReader(`{"id":"member@example.com","password":"Password!1"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if repo.migratedUID != 7 {
		t.Fatalf("migrated uid = %d", repo.migratedUID)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.migratedHash), []byte(password)); err != nil {
		t.Fatalf("password was not migrated to bcrypt: %v", err)
	}
}

func TestMobileRefreshRotatesAndReturnsTokenPair(t *testing.T) {
	previous := configs.Env
	configs.Env.JWTSecretKey = "test-secret"
	t.Cleanup(func() { configs.Env = previous })

	_, refreshDays := configs.GetJWTAccessRefresh()
	oldRefresh, err := utils.GenerateRefreshToken(7, refreshDays)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mobileRefreshRepo{userUid: 7, old: oldRefresh}
	authService := services.NewNuboAuthService(&repositories.Repository{Auth: repo})
	handler := NewNuboAuthHandler(&services.Service{Auth: authService})
	app := fiber.New()
	app.Post("/auth/android/refresh", handler.MobileRefreshAccessTokenHandler)
	body := strings.NewReader(`{"refresh":"` + oldRefresh + `"}`)
	req := httptest.NewRequest("POST", "/auth/android/refresh", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	var result struct {
		Success bool                 `json:"success"`
		Result  models.AuthTokenPair `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !result.Success || !repo.rotated || result.Result.Token == "" || result.Result.Refresh == "" {
		t.Fatalf("mobile refresh result = %#v, rotated = %v", result, repo.rotated)
	}
	if result.Result.Refresh == oldRefresh {
		t.Fatal("refresh token was not rotated")
	}
}

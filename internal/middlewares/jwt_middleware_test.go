package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/utils"
)

func TestJWTMiddlewareChecksCurrentAccountState(t *testing.T) {
	oldSecret := configs.Env.JWTSecretKey
	configs.Env.JWTSecretKey = "test-secret"
	t.Cleanup(func() { configs.Env.JWTSecretKey = oldSecret })

	token, err := utils.GenerateAccessToken(7, 1)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name       string
		allowed    bool
		wantStatus int
	}{
		{name: "active", allowed: true, wantStatus: fiber.StatusNoContent},
		{name: "blocked", allowed: false, wantStatus: fiber.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/protected", JWTMiddleware(func(userUid uint) bool {
				return userUid == 7 && tt.allowed
			}), func(c fiber.Ctx) error {
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAdminMiddlewareRequiresActiveRootAdministrator(t *testing.T) {
	oldSecret := configs.Env.JWTSecretKey
	configs.Env.JWTSecretKey = "test-secret"
	t.Cleanup(func() { configs.Env.JWTSecretKey = oldSecret })

	for _, tt := range []struct {
		name       string
		uid        uint
		active     bool
		wantStatus int
	}{
		{name: "active root", uid: 1, active: true, wantStatus: fiber.StatusNoContent},
		{name: "ordinary user", uid: 2, active: true, wantStatus: fiber.StatusOK},
		{name: "blocked root", uid: 1, active: false, wantStatus: fiber.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			token, err := utils.GenerateAccessToken(tt.uid, 1)
			if err != nil {
				t.Fatal(err)
			}
			called := false
			app := fiber.New()
			app.Get("/admin", AdminMiddleware(func(userUid uint) bool {
				return userUid == tt.uid && tt.active
			}), func(c fiber.Ctx) error {
				called = true
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if called != (tt.uid == 1 && tt.active) {
				t.Fatalf("protected handler called = %v", called)
			}
		})
	}
}

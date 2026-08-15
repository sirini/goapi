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

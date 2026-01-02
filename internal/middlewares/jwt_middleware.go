package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 로그인 여부를 확인하는 미들웨어
func JWTMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
		if actionUserUid < 1 {
			return utils.Err(c, "unauthorized: only logged-in users can operate functions", models.CODE_INVALID_TOKEN)
		}
		return c.Next()
	}
}

// 최고 관리자인지 확인하는 미들웨어
func AdminMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		actionUserUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
		if actionUserUid < 1 {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		if actionUserUid != 1 {
			return utils.Err(c, "unauthorized: you are not an administrator", models.CODE_NOT_ADMIN)
		}
		return c.Next()
	}
}

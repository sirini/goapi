package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/utils"
)

// 로그인 여부를 확인하는 미들웨어
func JWTMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		userUid := utils.ExtractUserUid(c.Get("Authorization"))
		if userUid < 1 {
			return utils.Err(c, "Unauthorized access, not a valid token")
		}
		return c.Next()
	}
}

// 최고 관리자인지 확인하는 미들웨어
func AdminMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		userUid := utils.ExtractUserUid(c.Get("Authorization"))
		if userUid != 1 {
			return utils.Err(c, "Unauthorized access, you are not an administrator")
		}
		return c.Next()
	}
}
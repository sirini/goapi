package utils

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
)

// 에러 메시지에 대한 응답
func Err(c fiber.Ctx, msg string) error {
	return c.JSON(models.ResponseCommon{
		Success: false,
		Error:   msg,
	})
}

// 성공 메시지 및 데이터 반환
func Ok(c fiber.Ctx, result interface{}) error {
	return c.JSON(models.ResponseCommon{
		Success: true,
		Result:  result,
	})
}
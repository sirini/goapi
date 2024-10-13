package handler

import "github.com/gofiber/fiber/v2"

func HelloHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Hello, Fiber! Hi Golang!",
	})
}
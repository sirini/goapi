package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirini/goapi/internal/handler"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/goapi")

	api.Get("/hello", handler.HelloHandler)
}
package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirini/goapi/internal/app/router"
	"github.com/sirini/goapi/internal/config"
)

// 서버 시작
func StartServer(cfg *config.Config) {
	app := fiber.New()
	router.SetupRoutes(app)

	port := fmt.Sprintf(":%s", cfg.Port)
	app.Listen(port)
}

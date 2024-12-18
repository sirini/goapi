package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/internal/routers"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
)

func main() {
	configs.LoadConfig()
	db := models.Connect(&configs.Env)
	defer db.Close()

	repo := repositories.NewRepository(db)
	service := services.NewService(repo)
	handler := handlers.NewHandler(service)

	app := fiber.New()
	goapi := app.Group("/goapi")
	routers.RegisterRouters(goapi, handler)

	port := fmt.Sprintf(":%s", configs.Env.Port)
	log.Printf("🚀 TSBOARD : GOAPI %v is running on %v", configs.Env.Version, configs.Env.Port)

	app.Listen(port)
}

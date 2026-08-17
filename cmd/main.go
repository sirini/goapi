package main

import (
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/internal/routers"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
)

func main() {
	if isInstalled := configs.Install(); !isInstalled {
		log.Fatalln("💣 Failed to install NUBO, the database connection details you provided may be incorrect ",
			"or you may not have the necessary permissions to create a new .env file. ",
			"Please leave a support request on the [nubohub.org] website!")
	}

	if err := configs.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	db := models.Connect(&configs.Env)
	defer db.Close()
	if len(os.Args) > 1 && os.Args[1] == "install" {
		if err := configs.InstallSchema(db, configs.Env.Prefix); err != nil {
			log.Fatalf("Failed to install database updates: %v", err)
		}
		log.Println("✅ Database updates installed")
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "update" {
		configs.Update(db, configs.Env.Prefix)
	}

	repo := repositories.NewRepository(db)
	service := services.NewService(repo)
	handler := handlers.NewHandler(service, db)

	sizeLimit := configs.GetFileSizeLimit()
	app := fiber.New(fiber.Config{
		BodyLimit: sizeLimit,
	})

	log.Printf("⚙️ Goapi base: %s\n", configs.Env.GoapiBase)
	log.Printf("⚙️ Domain: %s\n", configs.Env.Domain)
	log.Printf("⚙️ Title: %s\n", configs.Env.Title)
	log.Printf("⚙️ Listen: %s:%s\n", configs.Env.GoHost, configs.Env.GoPort)
	log.Printf("⚙️ Max body size: %d bytes", sizeLimit)

	goapi := app.Group(fmt.Sprintf("/%s", configs.Env.GoapiBase))
	routers.RegisterRouters(goapi, handler)

	address := net.JoinHostPort(configs.Env.GoHost, configs.Env.GoPort)
	log.Printf("🚀 GOAPI for NUBO %v is running on %v\n", configs.Env.Version, address)

	if err := app.Listen(address); err != nil {
		log.Printf("❌ Failed to start the goapi for NUBO: %v\n", err)
	}
}

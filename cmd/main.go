package main

import (
	"database/sql"
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
	if len(os.Args) > 1 && os.Args[1] == "install" {
		installDatabase()
		return
	}

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

// 외부 환경 파일을 기준으로 DB와 기본 데이터를 재실행 가능하게 준비한다.
func installDatabase() {
	if err := configs.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	db, err := openInstallDatabase()
	if err != nil {
		log.Fatalf("Failed to prepare database connection: %v", err)
	}
	defer db.Close()
	admin := configs.AdminInfo{Id: configs.Env.AdminID, Pw: configs.Env.AdminPW}
	if err := configs.BootstrapDatabase(db, configs.Env.Prefix, admin); err != nil {
		log.Fatalf("Failed to install database: %v", err)
	}
	log.Println("✅ Database installation completed")
}

// 기존 DB에는 그대로 연결하고, 없을 때만 서버 연결로 생성한 뒤 다시 연다.
func openInstallDatabase() (*sql.DB, error) {
	db, err := models.Open(&configs.Env, true)
	if err == nil {
		return db, nil
	}
	serverDB, serverErr := models.Open(&configs.Env, false)
	if serverErr != nil {
		return nil, serverErr
	}
	defer serverDB.Close()
	if createErr := configs.EnsureDatabase(serverDB, configs.Env.DBName); createErr != nil {
		return nil, createErr
	}
	return models.Open(&configs.Env, true)
}

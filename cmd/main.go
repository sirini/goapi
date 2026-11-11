package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/internal/routers"
	"github.com/sirini/goapi/internal/securitylog"
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

	days, parseErr := strconv.Atoi(configs.Env.SecurityLogDays)
	if parseErr != nil || days < 1 || days > 90 {
		log.Fatal("invalid NUBO_SECURITY_LOG_DAYS (1-90)")
	}
	maskedIP, parseErr := strconv.ParseBool(configs.Env.SecurityLogMaskedIP)
	if parseErr != nil {
		log.Fatal("invalid NUBO_SECURITY_LOG_MASKED_IP")
	}
	if err := securitylog.ValidatePrivateDirectory(configs.Env.SecurityLogDir, configs.Env.UploadDir); err != nil {
		log.Fatal("security log directory must be private and outside uploads")
	}
	var logErr error
	securitylog.Default, logErr = securitylog.New(configs.Env.SecurityLogDir, days, 10*1024*1024, maskedIP)
	if logErr != nil {
		log.Fatal("cannot initialize private security log directory")
	}
	auditDays, parseErr := strconv.Atoi(configs.Env.AuditLogDays)
	if parseErr != nil || auditDays < 367 || auditDays > 732 {
		log.Fatal("invalid NUBO_AUDIT_LOG_DAYS (367-732)")
	}
	securitylog.Audit, logErr = securitylog.NewAudit(filepath.Join(configs.Env.SecurityLogDir, "audit"), auditDays)
	if logErr != nil {
		log.Fatal("cannot initialize private administrator audit journal")
	}
	visitDays, visitErr := strconv.Atoi(configs.Env.VisitLogDays)
	if visitErr != nil || visitDays < 1 || visitDays > 732 {
		log.Fatal("invalid NUBO_VISIT_LOG_DAYS (1-732)")
	}
	if err := securitylog.ValidatePrivateDirectory(filepath.Join(configs.Env.SecurityLogDir, "visits"), configs.Env.UploadDir); err != nil {
		log.Fatal("visitor log directory must be private and outside uploads")
	}
	securitylog.Visits, logErr = securitylog.New(filepath.Join(configs.Env.SecurityLogDir, "visits"), visitDays, 100*1024*1024, false)
	if logErr != nil {
		log.Fatal("cannot initialize visitor logs")
	}
	securitylog.Visits.EnableFullIP()
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := securitylog.Visits.Prune(); err != nil {
				log.Print("visitor log expiry cleanup failed")
			}
			if err := securitylog.Default.Prune(); err != nil {
				log.Print("security log expiry cleanup failed")
			}
			if err := securitylog.Audit.Prune(); err != nil {
				log.Print("administrator audit expiry cleanup failed")
			}
		}
	}()
	app.Use(securitylog.Middleware)

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

package models

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
)

func Connect(cfg *configs.Config) *sql.DB {
	addr := fmt.Sprintf("tcp(%s:%s)", cfg.DBHost, cfg.DBPort)
	if len(cfg.DBSocket) > 0 {
		addr = cfg.DBSocket
	}
	log.Printf("🕑 Connect to the database by %s ...\n", addr)

	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s",
		cfg.DBUser, cfg.DBPass, addr, cfg.DBName, cfg.DBTimeout, cfg.DBReadTimeout, cfg.DBWriteTimeout)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Failed to connect to database: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("❌ Database ping failed: ", err)
	}

	maxIdle, err := strconv.ParseInt(cfg.DBMaxIdle, 10, 32)
	if err != nil {
		maxIdle = 100
	}
	maxOpen, err := strconv.ParseInt(cfg.DBMaxOpen, 10, 32)
	if err != nil {
		maxOpen = 100
	}

	db.SetMaxIdleConns(int(maxIdle))
	db.SetMaxOpenConns(int(maxOpen))
	db.SetConnMaxIdleTime(60 * time.Second)
	db.SetConnMaxLifetime(0)

	log.Printf(":: Max idle connections: %s\n", cfg.DBMaxIdle)
	log.Printf(":: Max open connections: %s\n", cfg.DBMaxOpen)
	log.Println("✅ Database connected successfully.")
	return db
}

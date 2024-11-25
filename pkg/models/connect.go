package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
)

func Connect(cfg *configs.Config) *sql.DB {
	addr := fmt.Sprintf("tcp(%s:%s)", cfg.DBHost, cfg.DBPort)
	if len(cfg.DBSocket) > 0 {
		addr = cfg.DBSocket
	}
	log.Printf("Connect to the database by %s ...\n", addr)

	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, addr, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database ping failed: ", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(60 * time.Second)
	db.SetConnMaxLifetime(0)

	log.Println("Database connected successfully.")
	return db
}

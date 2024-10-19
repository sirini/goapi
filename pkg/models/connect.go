package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirini/goapi/internal/configs"
)

func Connect(cfg *configs.Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database ping failed: ", err)
	}

	log.Println("Database connected successfully.")
	return db
}

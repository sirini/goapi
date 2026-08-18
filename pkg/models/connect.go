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

// 설정에 지정된 데이터베이스에 연결하며 실패 시 서버 시작을 중단한다.
func Connect(cfg *configs.Config) *sql.DB {
	db, err := Open(cfg, true)
	if err != nil {
		log.Fatal("🞬 Failed to connect to database: ", err)
	}
	return db
}

// 데이터베이스 생성 전에는 서버에만, 생성 후에는 지정한 DB까지 연결한다.
func Open(cfg *configs.Config, includeDatabase bool) (*sql.DB, error) {
	addr := fmt.Sprintf("tcp(%s:%s)", cfg.DBHost, cfg.DBPort)
	if len(cfg.DBSocket) > 0 {
		addr = fmt.Sprintf("unix(%s)", cfg.DBSocket)
	}
	log.Printf("🕑 Connect to the database by %s ...\n", addr)

	databaseName := ""
	if includeDatabase {
		databaseName = cfg.DBName
	}
	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&loc=Local",
		cfg.DBUser, cfg.DBPass, addr, databaseName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	maxIdle, err := strconv.ParseInt(cfg.DBMaxIdle, 10, 32)
	if err != nil {
		maxIdle = 10
	}
	maxOpen, err := strconv.ParseInt(cfg.DBMaxOpen, 10, 32)
	if err != nil {
		maxOpen = 10
	}

	db.SetMaxIdleConns(int(maxIdle))
	db.SetMaxOpenConns(int(maxOpen))
	db.SetConnMaxLifetime(3 * time.Minute)

	log.Printf("⚙️ Max idle connections: %s\n", cfg.DBMaxIdle)
	log.Printf("⚙️ Max open connections: %s\n", cfg.DBMaxOpen)
	log.Println("⚙️ Max lifetime of conn: 3 minutes")
	log.Println("✅ Database connected successfully, good to go!")
	return db, nil
}

package main

import (
	"log"

	"github.com/sirini/goapi/internal/app"
	"github.com/sirini/goapi/internal/config"
	"github.com/sirini/goapi/internal/model"
)

func main() {
	cfg := config.LoadConfig() // .env 설정 부르기
	model.Connect(cfg)         // DB에 연결하기

	log.Printf("🚀 TSBOARD : GOAPI is running on port %v\n", cfg.Port)
	app.StartServer(cfg) // 서버 리스닝 시작
}

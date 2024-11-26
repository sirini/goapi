package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/internal/routers"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
)

func main() {
	configs.LoadConfig()               // .env 설정 부르기
	db := models.Connect(&configs.Env) // DB에 연결하기
	defer db.Close()

	repo := repositories.NewRepository(db) // 리포지토리 등록하기
	service := services.NewService(repo)   // 서비스 등록하기

	mux := http.NewServeMux()
	routers.SetupRoutes(mux, service) // 라우터 등록하기

	log.Printf(`
  ___________ ____  ____  ___    ____  ____          __________ 
 /_  __/ ___// __ )/ __ \/   |  / __ \/ __ \   _    / ____/ __ \
  / /  \__ \/ __  / / / / /| | / /_/ / / / /  (_)  / / __/ / / /
 / /  ___/ / /_/ / /_/ / ___ |/ _, _/ /_/ /  _    / /_/ / /_/ /  
/_/  /____/_____/\____/_/  |_/_/ |_/_____/  (_)   \____/\____/
                                                                                  
🚀 TSBOARD %v is running on port %v [tsboard.dev]
	`, configs.Env.Version, configs.Env.Port)

	// 프로파일링
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	port := fmt.Sprintf(":%s", configs.Env.Port)
	log.Fatal(http.ListenAndServe(port, mux))
}

package main

import (
	"fmt"
	"log"
	"net/http"

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
  ___________ ____  ____  ___    ____  ____          __________  ___    ____  ____
 /_  __/ ___// __ )/ __ \/   |  / __ \/ __ \   _    / ____/ __ \/   |  / __ \/  _/
  / /  \__ \/ __  / / / / /| | / /_/ / / / /  (_)  / / __/ / / / /| | / /_/ // /  
 / /  ___/ / /_/ / /_/ / ___ |/ _, _/ /_/ /  _    / /_/ / /_/ / ___ |/ ____// /   
/_/  /____/_____/\____/_/  |_/_/ |_/_____/  (_)   \____/\____/_/  |_/_/   /___/   
                                                                                  
🚀 TSBOARD %v is running on port %v [tsboard.dev]
	`, configs.Env.Version, configs.Env.Port)

	port := fmt.Sprintf(":%s", configs.Env.Port)
	log.Fatal(http.ListenAndServe(port, mux))
}

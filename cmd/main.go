package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/sirini/goapi/internal/app"
	"github.com/sirini/goapi/internal/config"
	"github.com/sirini/goapi/internal/model"
)

func main() {
	mux := app.SetupRouter()
	cfg := config.LoadConfig() // .env 설정 부르기
	model.Connect(cfg)         // DB에 연결하기

	log.Printf(`
  ___________ ____  ____  ___    ____  ____          __________  ___    ____  ____
 /_  __/ ___// __ )/ __ \/   |  / __ \/ __ \   _    / ____/ __ \/   |  / __ \/  _/
  / /  \__ \/ __  / / / / /| | / /_/ / / / /  (_)  / / __/ / / / /| | / /_/ // /  
 / /  ___/ / /_/ / /_/ / ___ |/ _, _/ /_/ /  _    / /_/ / /_/ / ___ |/ ____// /   
/_/  /____/_____/\____/_/  |_/_/ |_/_____/  (_)   \____/\____/_/  |_/_/   /___/   
                                                                                  
🚀 TSBOARD %v is running on port %v [tsboard.dev]
	`, cfg.Version, cfg.Port)

	var builder strings.Builder
	builder.WriteString(":")
	builder.WriteString(cfg.Port)
	port := builder.String()

	http.ListenAndServe(port, mux)
}

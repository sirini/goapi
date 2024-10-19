package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sirini/goapi/internal/configs"
)

// 응답 구조체
type ResponseVisit struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Version string `json:"version"`
	License string `json:"license"`
	Github  string `json:"github"`
}

// 메세지 출력 테스트
func Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := ResponseVisit{
		Success: true,
		Message: "Thanks for using the TSBOARD!",
		Version: configs.Env.Version,
		License: "MIT",
		Github:  "github.com/sirini/goapi",
	}

	json.NewEncoder(w).Encode(response)
}

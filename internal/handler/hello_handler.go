package handler

import (
	"encoding/json"
	"net/http"
)

// 응답 구조체
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// 메세지 출력 테스트
func Hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: true,
		Message: "Welcome!",
	}
	json.NewEncoder(w).Encode(response)
}

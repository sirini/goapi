package utils

import (
	"encoding/json"
	"net/http"
)

// JSON 응답을 처리하는 헬퍼
func ResponseJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

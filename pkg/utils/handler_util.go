package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirini/goapi/pkg/models"
)

// JSON 응답을 처리하는 헬퍼
func ResponseJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// 에러 메시지 JSON 응답 처리용 헬퍼
func ResponseError(w http.ResponseWriter, message string) {
	ResponseJSON(w, models.ResponseCommon{
		Success: false,
		Error:   message,
	})
}

// 성공 메시지 JSON 응답 처리용 헬퍼
func ResponseSuccess(w http.ResponseWriter, result interface{}) {
	ResponseJSON(w, models.ResponseCommon{
		Success: true,
		Result:  result,
	})
}

package middlewares

import (
	"net/http"

	"github.com/sirini/goapi/pkg/utils"
)

// 에러 시 응답
type response struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// 로그인 여부를 확인하는 미들웨어
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			utils.ResponseJSON(w, response{
				Success: false,
				Error:   "Unauthorized access: no token provided",
			})
		}
		// TODO
	})
}

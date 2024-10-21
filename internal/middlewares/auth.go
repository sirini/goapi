package middlewares

import (
	"net/http"
	"strings"

	"github.com/sirini/goapi/pkg/utils"
)

// 로그인 여부를 확인하는 미들웨어
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			utils.ResponseError(w, "Unauthorized access: no token provided")
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ResponseError(w, "Invalid authorization format")
			return
		}

		_, err := utils.ValidateJWT(parts[1])
		if err != nil {
			utils.ResponseError(w, "Invalid or expired token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

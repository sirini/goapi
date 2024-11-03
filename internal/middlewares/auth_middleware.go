package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 로그인 여부를 확인하는 미들웨어
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			utils.ResponseError(w, "Unauthorized access: no token provided")
			return
		}

		parts := strings.Split(tokenStr, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ResponseError(w, "Invalid authorization format")
			return
		}

		token, err := utils.ValidateJWT(parts[1])
		if err != nil {
			utils.ResponseError(w, err.Error())
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.ResponseError(w, "Unable to extract claims")
			return
		}

		ctx := context.WithValue(r.Context(), models.JwtClaimsKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

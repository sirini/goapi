package middlewares

import (
	"context"
	"net/http"

	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 로그인 여부를 확인하는 미들웨어
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userUid := utils.FindUserUidFromHeader(r)
		if userUid < 1 {
			utils.ResponseError(w, "Unauthorized access, not a valid token")
			return
		}

		ctx := context.WithValue(r.Context(), models.JwtClaimsKey, userUid)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

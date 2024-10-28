package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/internal/configs"
)

// JWT 토큰 검증
func ValidateJWT(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(configs.Env.JWTSecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

// 아이디가 이메일 형식에 부합하는지 확인
func IsValidEmail(email string) bool {
	const regexPattern = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(regexPattern)
	return re.MatchString(email)
}

// 주어진 문자열을 sha256 알고리즘으로 변환
func GetHashedString(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// 리프레시 토큰을 쿠키에 저장
func SetRefreshCookie(w http.ResponseWriter, refreshToken string) {
	cookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   86400 * 14, /* days */
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

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
func SaveCookie(w http.ResponseWriter, name string, value string, days int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   86400 * days,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// 액세스 토큰 생성하기
func GenerateAccessToken(userUid uint, hours time.Duration) (string, error) {
	auth := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": userUid,
		"exp": time.Now().Add(time.Hour * hours).Unix(),
	})
	return auth.SignedString([]byte(configs.Env.JWTSecretKey))
}

// 리프레시 토큰 생성하기
func GenerateRefreshToken(months int) (string, error) {
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().AddDate(0, months, 0).Unix(),
	})
	return refresh.SignedString([]byte(configs.Env.JWTSecretKey))
}

// 구조체를 JSON 형식의 문자열로 변환
func ConvertJsonString(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(data)
	return encoded, nil
}

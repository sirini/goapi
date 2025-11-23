package utils

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
	"golang.org/x/oauth2"
)

// 구조체를 JSON 형식의 문자열로 변환
func ConvJsonString(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(data)
	return encoded, nil
}

// 주어진 문자열을 sha256 알고리즘으로 변환
func GetHashedString(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// 액세스 토큰 생성하기 (유효시간 기입 필요)
func GenerateAccessToken(userUid uint, hours int) (string, error) {
	auth := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": userUid,
		"exp": time.Now().Add(time.Hour * time.Duration(hours)).Unix(),
	})
	return auth.SignedString([]byte(configs.Env.JWTSecretKey))
}

// 리프레시 토큰 생성하기 (유효일자 기입 필요)
func GenerateRefreshToken(days int) (string, error) {
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().AddDate(0, 0, days).Unix(),
	})
	return refresh.SignedString([]byte(configs.Env.JWTSecretKey))
}

// 헤더로 넘어온 Authorization 문자열 추출해서 사용자 고유 번호 반환
func ExtractUserUid(authorization string) int {
	if authorization == "" {
		return models.JWT_EMPTY_TOKEN
	}
	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return models.JWT_NOT_BEARER
	}
	token, err := ValidateJWT(parts[1])
	if err != nil {
		return models.JWT_INVALID_TOKEN
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return models.JWT_NO_CLAIMS
	}
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return models.JWT_NO_UID
	}
	return int(uidFloat)
}

// 아이디가 이메일 형식에 부합하는지 확인
func IsValidEmail(email string) bool {
	const regexPattern = `^(?i)[a-z0-9._%+\-]+@[a-z0-9\-]+(\.[a-z0-9\-]+)*\.[a-z]{2,}$`
	re := regexp.MustCompile(regexPattern)
	return re.MatchString(email)
}

// 쿠키에 저장
func SaveCookie(c fiber.Ctx, name string, value string, days int) {
	isSecure := false
	if strings.HasPrefix(configs.Env.URL, "https://") {
		isSecure = true
	}

	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   86400 * days,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   isSecure,
	})
}

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

// 상태 검사 및 토큰 교환 후 토큰 반환
func OAuth2ExchangeToken(c fiber.Ctx, cfg oauth2.Config) (*oauth2.Token, error) {
	cookie := c.Cookies("nubo-oauth-state")
	if cookie != c.FormValue("state") {
		c.Redirect().To(configs.Env.URL)
		return nil, fmt.Errorf("empty oauth state from cookie")
	}

	code := c.FormValue("code")
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		c.Redirect().To(configs.Env.URL)
		return nil, err
	}
	return token, nil
}

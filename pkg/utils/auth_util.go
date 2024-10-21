package utils

import (
	"fmt"
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
		return configs.Env.JWTSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

// JWT 토큰 생성
func GenerateJWT(userUid uint) (string, string, error) {
	auth := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": userUid,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().AddDate(0, 1, 0).Unix(),
	})

	authToken, err := auth.SignedString(configs.Env.JWTSecretKey)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := refresh.SignedString(configs.Env.JWTSecretKey)
	if err != nil {
		return authToken, "", err
	}

	return authToken, refreshToken, nil
}

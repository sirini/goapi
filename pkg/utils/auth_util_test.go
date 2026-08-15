package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/internal/configs"
)

func TestGenerateRefreshTokenIsUnique(t *testing.T) {
	first, err := GenerateRefreshToken(1, 30)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() first call error = %v", err)
	}
	second, err := GenerateRefreshToken(1, 30)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() second call error = %v", err)
	}
	if first == second {
		t.Fatal("refresh tokens generated in the same second must be unique")
	}
}

func TestValidateJWTRejectsUnexpectedHMACAlgorithm(t *testing.T) {
	oldSecret := configs.Env.JWTSecretKey
	configs.Env.JWTSecretKey = "test-secret"
	t.Cleanup(func() { configs.Env.JWTSecretKey = oldSecret })

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"uid": 1,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(configs.Env.JWTSecretKey))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateJWT(signed); err == nil {
		t.Fatal("HS512 token was accepted by an HS256-only validator")
	}
}

func TestOAuthStateMatchesRejectsEmptyValues(t *testing.T) {
	if OAuthStateMatches("", "") {
		t.Fatal("empty OAuth state values matched")
	}
	if OAuthStateMatches("state", "") || OAuthStateMatches("", "state") {
		t.Fatal("partially empty OAuth state matched")
	}
	if !OAuthStateMatches("state", "state") {
		t.Fatal("equal non-empty OAuth state did not match")
	}
}

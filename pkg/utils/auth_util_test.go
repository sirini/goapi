package utils

import "testing"

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

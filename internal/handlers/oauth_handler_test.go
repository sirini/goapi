package handlers

import (
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

func TestValidGoogleIDTokenInfo(t *testing.T) {
	valid := models.GoogleUser{
		Audience:      "android-client-id",
		Email:         "user@example.com",
		EmailVerified: "true",
	}
	if !validGoogleIDTokenInfo(valid, "android-client-id") {
		t.Fatal("valid Google ID token claims were rejected")
	}

	wrongAudience := valid
	wrongAudience.Audience = "attacker-client-id"
	if validGoogleIDTokenInfo(wrongAudience, "android-client-id") {
		t.Fatal("Google ID token for a different audience was accepted")
	}

	unverified := valid
	unverified.EmailVerified = "false"
	if validGoogleIDTokenInfo(unverified, "android-client-id") {
		t.Fatal("Google ID token with an unverified email was accepted")
	}
}

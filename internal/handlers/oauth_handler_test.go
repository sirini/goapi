package handlers

import (
	"errors"
	"testing"

	"github.com/sirini/goapi/pkg/models"
	"google.golang.org/api/idtoken"
)

func TestPublicAppleVerificationErrorOnlyExposesSafeCategories(t *testing.T) {
	for _, message := range []string{
		"Apple token audience is not allowed",
		"Apple token nonce does not match",
		"Apple email is not verified",
	} {
		if got := publicAppleVerificationError(errors.New(message)); got != message {
			t.Fatalf("public Apple error = %q, want %q", got, message)
		}
	}
	if got := publicAppleVerificationError(errors.New("dial tcp 10.0.0.1: secret")); got != "invalid Apple identity token" {
		t.Fatalf("internal verifier error was exposed: %q", got)
	}
}

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

func TestGoogleUserFromIDTokenPayload(t *testing.T) {
	user := googleUserFromIDTokenPayload(&idtoken.Payload{
		Audience: "mobile-server-client-id",
		Subject:  "google-subject",
		Claims: map[string]interface{}{
			"email":          "user@example.com",
			"email_verified": true,
			"name":           "Google User",
			"picture":        "https://example.com/profile.jpg",
		},
	})
	if user.ID != "google-subject" || user.Audience != "mobile-server-client-id" ||
		user.Email != "user@example.com" || user.EmailVerified != "true" ||
		user.Name != "Google User" || user.Picture != "https://example.com/profile.jpg" {
		t.Fatalf("unexpected Google payload conversion: %+v", user)
	}

	invalid := googleUserFromIDTokenPayload(&idtoken.Payload{
		Audience: "mobile-server-client-id",
		Claims: map[string]interface{}{
			"email":          "user@example.com",
			"email_verified": "true",
		},
	})
	if validGoogleIDTokenInfo(invalid, "mobile-server-client-id") {
		t.Fatal("non-boolean email verification claim was accepted")
	}
	if user := googleUserFromIDTokenPayload(nil); user.Email != "" {
		t.Fatal("nil payload produced a populated Google user")
	}
}

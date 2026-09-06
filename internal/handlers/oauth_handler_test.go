package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"google.golang.org/api/idtoken"
)

type appleDeleteOAuthStub struct {
	services.OAuthService
	linkedUser    uint
	nonceConsumed bool
}

func (s *appleDeleteOAuthStub) FindAppleUser(subject string) (uint, bool, error) {
	return s.linkedUser, subject == "apple-subject" && s.linkedUser > 0, nil
}

func (s *appleDeleteOAuthStub) ConsumeAppleNonce(purpose string, userUid uint, nonce string) (bool, error) {
	s.nonceConsumed = purpose == appleNonceDelete && userUid == 7 && nonce == "delete-nonce"
	return s.nonceConsumed, nil
}

type appleDeleteUserStub struct {
	services.UserService
	deleted bool
}

func (s *appleDeleteUserStub) DeleteAccount(userUid uint, confirmation string) error {
	s.deleted = userUid == 7 && confirmation == "DELETE"
	return nil
}

type appleDeleteVerifierStub struct{}

func (appleDeleteVerifierStub) Verify(_ context.Context, token, nonce string, audiences []string) (models.AppleIdentity, error) {
	if nonce != "delete-nonce" || len(audiences) == 0 || audiences[0] != "me.sensta.ios" {
		return models.AppleIdentity{}, errors.New("invalid test verification")
	}
	if token != "fresh.identity" && token != "exchanged.identity" {
		return models.AppleIdentity{}, errors.New("invalid test token")
	}
	return models.AppleIdentity{Subject: "apple-subject", Audience: "me.sensta.ios"}, nil
}

type appleDeleteRevokerStub struct {
	failRevoke bool
	exchanged  bool
	revoked    bool
}

func (*appleDeleteRevokerStub) Configured() bool { return true }
func (s *appleDeleteRevokerStub) Exchange(_ context.Context, clientID, code string) (services.AppleTokenExchange, error) {
	s.exchanged = clientID == "me.sensta.ios" && code == "fresh-code"
	return services.AppleTokenExchange{
		IdentityToken: "exchanged.identity", RefreshToken: "apple-refresh",
	}, nil
}
func (s *appleDeleteRevokerStub) Revoke(_ context.Context, clientID, refresh string) error {
	if s.failRevoke {
		return errors.New("Apple unavailable")
	}
	s.revoked = clientID == "me.sensta.ios" && refresh == "apple-refresh"
	return nil
}

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

func TestAppleAccountDeletionRevokesBeforeDeletingLocalAccount(t *testing.T) {
	previous := configs.Env
	configs.Env.JWTSecretKey = "test-secret"
	configs.Env.OAuthAppleClientIDs = "me.sensta.ios"
	t.Cleanup(func() { configs.Env = previous })
	oauth := &appleDeleteOAuthStub{linkedUser: 7}
	user := &appleDeleteUserStub{}
	revoker := &appleDeleteRevokerStub{}
	handler := &NuboOAuth2Handler{
		service:       &services.Service{OAuth: oauth, User: user},
		appleVerifier: appleDeleteVerifierStub{}, appleRevoker: revoker,
	}

	response := appleDeleteRequest(t, handler)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if !oauth.nonceConsumed || !revoker.exchanged || !revoker.revoked || !user.deleted {
		t.Fatalf("deletion sequence incomplete: nonce=%v exchange=%v revoke=%v delete=%v", oauth.nonceConsumed, revoker.exchanged, revoker.revoked, user.deleted)
	}
}

func TestAppleAccountDeletionKeepsLocalAccountWhenRevocationFails(t *testing.T) {
	previous := configs.Env
	configs.Env.JWTSecretKey = "test-secret"
	configs.Env.OAuthAppleClientIDs = "me.sensta.ios"
	t.Cleanup(func() { configs.Env = previous })
	user := &appleDeleteUserStub{}
	handler := &NuboOAuth2Handler{
		service:       &services.Service{OAuth: &appleDeleteOAuthStub{linkedUser: 7}, User: user},
		appleVerifier: appleDeleteVerifierStub{},
		appleRevoker:  &appleDeleteRevokerStub{failRevoke: true},
	}

	response := appleDeleteRequest(t, handler)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if user.deleted {
		t.Fatal("local account was deleted after Apple revocation failed")
	}
}

func appleDeleteRequest(t *testing.T, handler *NuboOAuth2Handler) *http.Response {
	t.Helper()
	token, err := utils.GenerateAccessToken(7, 1)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Delete("/apple/account", handler.AppleDeleteAccountHandler)
	request := httptest.NewRequest("DELETE", "/apple/account", strings.NewReader(
		`{"identityToken":"fresh.identity","authorizationCode":"fresh-code","nonce":"delete-nonce","confirmation":"DELETE"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(models.AUTH_KEY, "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
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

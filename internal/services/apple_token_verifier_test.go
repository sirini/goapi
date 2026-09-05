package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAppleTokenVerifierValidatesSignatureClaimsAndNonce(t *testing.T) {
	privateKey := appleTestKey(t)
	server := appleKeyServer(t, map[string]*rsa.PublicKey{"apple-key": &privateKey.PublicKey}, nil)
	defer server.Close()

	verifier := NewAppleTokenVerifier()
	verifier.keysURL = server.URL
	validClaims := jwt.MapClaims{
		"iss":            appleIssuer,
		"aud":            "me.sensta.ios",
		"sub":            "apple-subject",
		"nonce":          "server-nonce",
		"email":          "person@example.com",
		"email_verified": true,
		"iat":            time.Now().Add(-time.Minute).Unix(),
		"exp":            time.Now().Add(5 * time.Minute).Unix(),
	}

	identity, err := verifier.Verify(context.Background(), appleTestToken(t, privateKey, "apple-key", validClaims), "server-nonce", []string{"me.sensta.ios.debug", "me.sensta.ios"})
	if err != nil {
		t.Fatal(err)
	}
	if identity.Subject != "apple-subject" || identity.Email != "person@example.com" || !identity.EmailVerified || identity.Audience != "me.sensta.ios" {
		t.Fatalf("unexpected verified identity: %+v", identity)
	}

	for _, test := range []struct {
		name      string
		claims    jwt.MapClaims
		nonce     string
		audiences []string
	}{
		{name: "wrong nonce", claims: cloneClaims(validClaims), nonce: "other-nonce", audiences: []string{"me.sensta.ios"}},
		{name: "wrong audience", claims: cloneClaims(validClaims), nonce: "server-nonce", audiences: []string{"attacker.app"}},
		{name: "unverified email", claims: withClaim(validClaims, "email_verified", false), nonce: "server-nonce", audiences: []string{"me.sensta.ios"}},
		{name: "wrong issuer", claims: withClaim(validClaims, "iss", "https://attacker.example"), nonce: "server-nonce", audiences: []string{"me.sensta.ios"}},
		{name: "expired", claims: withClaim(validClaims, "exp", time.Now().Add(-time.Minute).Unix()), nonce: "server-nonce", audiences: []string{"me.sensta.ios"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			token := appleTestToken(t, privateKey, "apple-key", test.claims)
			if _, err := verifier.Verify(context.Background(), token, test.nonce, test.audiences); err == nil {
				t.Fatal("invalid Apple token was accepted")
			}
		})
	}
}

func TestAppleTokenVerifierRefreshesKeysOnUnknownKeyID(t *testing.T) {
	first := appleTestKey(t)
	rotated := appleTestKey(t)
	var calls atomic.Int32
	server := appleKeyServer(t, nil, func() map[string]*rsa.PublicKey {
		if calls.Add(1) == 1 {
			return map[string]*rsa.PublicKey{"first": &first.PublicKey}
		}
		return map[string]*rsa.PublicKey{"rotated": &rotated.PublicKey}
	})
	defer server.Close()
	verifier := NewAppleTokenVerifier()
	verifier.keysURL = server.URL

	claims := jwt.MapClaims{
		"iss": appleIssuer, "aud": "me.sensta.ios", "sub": "subject", "nonce": "nonce",
		"iat": time.Now().Add(-time.Minute).Unix(), "exp": time.Now().Add(5 * time.Minute).Unix(),
	}
	if _, err := verifier.Verify(context.Background(), appleTestToken(t, first, "first", claims), "nonce", []string{"me.sensta.ios"}); err != nil {
		t.Fatal(err)
	}
	if _, err := verifier.Verify(context.Background(), appleTestToken(t, rotated, "rotated", claims), "nonce", []string{"me.sensta.ios"}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("public key requests = %d, want 2", calls.Load())
	}
}

func appleTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func appleKeyServer(t *testing.T, static map[string]*rsa.PublicKey, dynamic func() map[string]*rsa.PublicKey) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		keys := static
		if dynamic != nil {
			keys = dynamic()
		}
		document := appleJWKS{Keys: make([]appleJWK, 0, len(keys))}
		for keyID, key := range keys {
			document.Keys = append(document.Keys, appleJWK{
				KeyType: "RSA", KeyID: keyID, Use: "sig", Algorithm: "RS256",
				Modulus:  base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
				Exponent: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
			})
		}
		if err := json.NewEncoder(w).Encode(document); err != nil {
			t.Error(err)
		}
	}))
}

func appleTestToken(t *testing.T, key *rsa.PrivateKey, keyID string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func cloneClaims(source jwt.MapClaims) jwt.MapClaims {
	result := jwt.MapClaims{}
	for key, value := range source {
		result[key] = value
	}
	return result
}

func withClaim(source jwt.MapClaims, key string, value any) jwt.MapClaims {
	result := cloneClaims(source)
	result[key] = value
	return result
}

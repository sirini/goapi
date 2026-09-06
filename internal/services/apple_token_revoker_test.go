package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAppleAuthorizationRevokerExchangesAndRevokesRefreshToken(t *testing.T) {
	privateKey, pemText := appleClientSecretTestKey(t)
	fixedNow := time.Unix(1_800_000_000, 0)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if request.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Fatalf("content type = %q", request.Header.Get("Content-Type"))
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(request.Form.Get("client_secret"), claims, func(token *jwt.Token) (any, error) {
			if token.Header["kid"] != "APPLEKEY" || token.Method.Alg() != "ES256" {
				t.Fatalf("unexpected client secret header: %#v", token.Header)
			}
			return &privateKey.PublicKey, nil
		}, jwt.WithValidMethods([]string{"ES256"}), jwt.WithAudience(appleIssuer), jwt.WithIssuer("TEAM123"))
		if err != nil || !token.Valid {
			t.Fatalf("invalid Apple client secret: %v", err)
		}
		if claims["sub"] != "me.sensta.ios" {
			t.Fatalf("client secret subject = %#v", claims["sub"])
		}
		switch request.URL.Path {
		case "/auth/token":
			if request.Form.Get("code") != "fresh-code" || request.Form.Get("grant_type") != "authorization_code" {
				t.Fatalf("unexpected token form: %#v", request.Form)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id_token": "exchanged.id.token", "refresh_token": "apple-refresh-token",
			})
		case "/auth/revoke":
			if request.Form.Get("token") != "apple-refresh-token" || request.Form.Get("token_type_hint") != "refresh_token" {
				t.Fatalf("unexpected revoke form: %#v", request.Form)
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	revoker := &AppleAuthorizationRevoker{
		client: server.Client(), tokenURL: server.URL + "/auth/token",
		revokeURL: server.URL + "/auth/revoke", teamID: "TEAM123", keyID: "APPLEKEY",
		privateKeyPEM: pemText, now: func() time.Time { return fixedNow },
	}
	exchange, err := revoker.Exchange(context.Background(), "me.sensta.ios", "fresh-code")
	if err != nil {
		t.Fatal(err)
	}
	if exchange.IdentityToken != "exchanged.id.token" || exchange.RefreshToken != "apple-refresh-token" {
		t.Fatalf("unexpected exchange: %+v", exchange)
	}
	if err := revoker.Revoke(context.Background(), "me.sensta.ios", exchange.RefreshToken); err != nil {
		t.Fatal(err)
	}
}

func TestAppleAuthorizationRevokerRejectsFailedExchange(t *testing.T) {
	_, pemText := appleClientSecretTestKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()
	revoker := &AppleAuthorizationRevoker{
		client: server.Client(), tokenURL: server.URL, teamID: "TEAM123", keyID: "APPLEKEY",
		privateKeyPEM: pemText, now: time.Now,
	}
	if _, err := revoker.Exchange(context.Background(), "me.sensta.ios", "expired-code"); err == nil {
		t.Fatal("failed Apple authorization exchange was accepted")
	}
}

func appleClientSecretTestKey(t *testing.T) (*ecdsa.PrivateKey, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return key, string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}))
}

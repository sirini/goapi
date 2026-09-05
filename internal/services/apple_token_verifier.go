package services

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/pkg/models"
)

const appleIssuer = "https://appleid.apple.com"
const applePublicKeysURL = "https://appleid.apple.com/auth/keys"

type AppleTokenVerifying interface {
	Verify(ctx context.Context, identityToken, nonce string, audiences []string) (models.AppleIdentity, error)
}

type appleJWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

type AppleTokenVerifier struct {
	client    *http.Client
	keysURL   string
	cacheTTL  time.Duration
	now       func() time.Time
	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func NewAppleTokenVerifier() *AppleTokenVerifier {
	return &AppleTokenVerifier{
		client:   &http.Client{Timeout: 5 * time.Second},
		keysURL:  applePublicKeysURL,
		cacheTTL: 6 * time.Hour,
		now:      time.Now,
	}
}

// Verify는 Apple 서명, issuer, audience, 만료, nonce와 이메일 검증 상태를 확인한다.
func (v *AppleTokenVerifier) Verify(ctx context.Context, identityToken, nonce string, audiences []string) (models.AppleIdentity, error) {
	if strings.TrimSpace(identityToken) == "" || strings.TrimSpace(nonce) == "" || len(audiences) == 0 {
		return models.AppleIdentity{}, errors.New("missing Apple token verification input")
	}
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(appleIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(30*time.Second),
	)
	token, err := parser.ParseWithClaims(identityToken, claims, func(token *jwt.Token) (any, error) {
		keyID, _ := token.Header["kid"].(string)
		if keyID == "" {
			return nil, errors.New("Apple token has no key id")
		}
		return v.publicKey(ctx, keyID)
	})
	if err != nil {
		return models.AppleIdentity{}, fmt.Errorf("invalid Apple identity token: %w", err)
	}
	if !token.Valid {
		return models.AppleIdentity{}, errors.New("invalid Apple identity token")
	}

	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		return models.AppleIdentity{}, errors.New("Apple token has no subject")
	}
	issuedAt, err := claims.GetIssuedAt()
	if err != nil || issuedAt == nil {
		return models.AppleIdentity{}, errors.New("Apple token has no issued-at time")
	}
	tokenAudiences, err := claims.GetAudience()
	matchedAudience := matchingAudience(tokenAudiences, audiences)
	if err != nil || matchedAudience == "" {
		return models.AppleIdentity{}, errors.New("Apple token audience is not allowed")
	}
	claimNonce, _ := claims["nonce"].(string)
	if claimNonce == "" || subtle.ConstantTimeCompare([]byte(claimNonce), []byte(nonce)) != 1 {
		return models.AppleIdentity{}, errors.New("Apple token nonce does not match")
	}
	email, _ := claims["email"].(string)
	email = strings.TrimSpace(email)
	emailVerified := appleBooleanClaim(claims["email_verified"])
	if email != "" && !emailVerified {
		return models.AppleIdentity{}, errors.New("Apple email is not verified")
	}
	return models.AppleIdentity{
		Subject:       subject,
		Audience:      matchedAudience,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}

func (v *AppleTokenVerifier) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if key := v.keys[keyID]; key != nil && v.now().Sub(v.fetchedAt) < v.cacheTTL {
		return key, nil
	}
	keys, err := v.fetchKeys(ctx)
	if err != nil {
		return nil, err
	}
	v.keys = keys
	v.fetchedAt = v.now()
	key := keys[keyID]
	if key == nil {
		return nil, errors.New("Apple signing key was not found")
	}
	return key, nil
}

func (v *AppleTokenVerifier) fetchKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.keysURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Apple public keys returned status %d", resp.StatusCode)
	}
	var document appleJWKS
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, key := range document.Keys {
		if key.KeyType != "RSA" || key.Algorithm != "RS256" || key.Use != "sig" || key.KeyID == "" {
			continue
		}
		publicKey, err := rsaPublicKey(key.Modulus, key.Exponent)
		if err != nil {
			continue
		}
		keys[key.KeyID] = publicKey
	}
	if len(keys) == 0 {
		return nil, errors.New("Apple public key response contained no supported keys")
	}
	return keys, nil
}

func rsaPublicKey(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(modulus)
	if err != nil || len(nBytes) == 0 {
		return nil, errors.New("invalid RSA modulus")
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(exponent)
	if err != nil || len(eBytes) == 0 || len(eBytes) > 4 {
		return nil, errors.New("invalid RSA exponent")
	}
	e := new(big.Int).SetBytes(eBytes).Int64()
	if e < 3 || e > int64(^uint(0)>>1) {
		return nil, errors.New("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(e)}, nil
}

func matchingAudience(tokenAudiences, allowed []string) string {
	for _, tokenAudience := range tokenAudiences {
		for _, allowedAudience := range allowed {
			if tokenAudience == allowedAudience && tokenAudience != "" {
				return tokenAudience
			}
		}
	}
	return ""
}

func appleBooleanClaim(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(typed, "true")
	default:
		return false
	}
}

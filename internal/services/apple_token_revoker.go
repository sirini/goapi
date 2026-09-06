package services

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirini/goapi/internal/configs"
)

const appleTokenURL = "https://appleid.apple.com/auth/token"
const appleRevokeURL = "https://appleid.apple.com/auth/revoke"

type AppleTokenExchange struct {
	IdentityToken string
	RefreshToken  string
}

type AppleAuthorizationRevoking interface {
	Configured() bool
	Exchange(ctx context.Context, clientID, authorizationCode string) (AppleTokenExchange, error)
	Revoke(ctx context.Context, clientID, refreshToken string) error
}

type AppleAuthorizationRevoker struct {
	client         *http.Client
	tokenURL       string
	revokeURL      string
	teamID         string
	keyID          string
	privateKeyPEM  string
	privateKeyFile string
	now            func() time.Time
	keyMu          sync.Mutex
	key            *ecdsa.PrivateKey
}

func NewAppleAuthorizationRevoker() *AppleAuthorizationRevoker {
	return &AppleAuthorizationRevoker{
		client:         &http.Client{Timeout: 10 * time.Second},
		tokenURL:       appleTokenURL,
		revokeURL:      appleRevokeURL,
		teamID:         strings.TrimSpace(configs.Env.OAuthAppleTeamID),
		keyID:          strings.TrimSpace(configs.Env.OAuthAppleKeyID),
		privateKeyPEM:  strings.TrimSpace(configs.Env.OAuthApplePrivateKey),
		privateKeyFile: strings.TrimSpace(configs.Env.OAuthApplePrivateKeyFile),
		now:            time.Now,
	}
}

func (r *AppleAuthorizationRevoker) Configured() bool {
	return r != nil && r.teamID != "" && r.keyID != "" && (r.privateKeyPEM != "" || r.privateKeyFile != "")
}

// Exchange는 5분짜리 일회용 authorization code를 서버 토큰으로 교환한다.
func (r *AppleAuthorizationRevoker) Exchange(ctx context.Context, clientID, authorizationCode string) (AppleTokenExchange, error) {
	clientID = strings.TrimSpace(clientID)
	authorizationCode = strings.TrimSpace(authorizationCode)
	if !r.Configured() || clientID == "" || authorizationCode == "" {
		return AppleTokenExchange{}, errors.New("Apple token revocation is not configured")
	}
	secret, err := r.clientSecret(clientID)
	if err != nil {
		return AppleTokenExchange{}, err
	}
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {secret},
		"code":          {authorizationCode},
		"grant_type":    {"authorization_code"},
	}
	response, err := r.postForm(ctx, r.tokenURL, form)
	if err != nil {
		return AppleTokenExchange{}, fmt.Errorf("Apple authorization code exchange failed: %w", err)
	}
	var token struct {
		IdentityToken string `json:"id_token"`
		RefreshToken  string `json:"refresh_token"`
	}
	if err := json.Unmarshal(response, &token); err != nil || token.IdentityToken == "" || token.RefreshToken == "" {
		return AppleTokenExchange{}, errors.New("Apple authorization code exchange returned an invalid response")
	}
	return AppleTokenExchange{IdentityToken: token.IdentityToken, RefreshToken: token.RefreshToken}, nil
}

// Revoke는 계정 삭제 전에 Apple refresh token과 해당 사용자 승인을 폐기한다.
func (r *AppleAuthorizationRevoker) Revoke(ctx context.Context, clientID, refreshToken string) error {
	clientID = strings.TrimSpace(clientID)
	refreshToken = strings.TrimSpace(refreshToken)
	if !r.Configured() || clientID == "" || refreshToken == "" {
		return errors.New("Apple token revocation is not configured")
	}
	secret, err := r.clientSecret(clientID)
	if err != nil {
		return err
	}
	_, err = r.postForm(ctx, r.revokeURL, url.Values{
		"client_id":       {clientID},
		"client_secret":   {secret},
		"token":           {refreshToken},
		"token_type_hint": {"refresh_token"},
	})
	if err != nil {
		return fmt.Errorf("Apple refresh token revocation failed: %w", err)
	}
	return nil
}

func (r *AppleAuthorizationRevoker) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := r.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Apple endpoint returned status %d", response.StatusCode)
	}
	return body, nil
}

func (r *AppleAuthorizationRevoker) clientSecret(clientID string) (string, error) {
	key, err := r.privateKey()
	if err != nil {
		return "", err
	}
	now := r.now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": r.teamID,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": appleIssuer,
		"sub": clientID,
	})
	token.Header["kid"] = r.keyID
	return token.SignedString(key)
}

func (r *AppleAuthorizationRevoker) privateKey() (*ecdsa.PrivateKey, error) {
	r.keyMu.Lock()
	defer r.keyMu.Unlock()
	if r.key != nil {
		return r.key, nil
	}
	pemText := r.privateKeyPEM
	if pemText == "" {
		info, err := os.Stat(r.privateKeyFile)
		if err != nil || info.Size() > 64*1024 {
			return nil, errors.New("failed to load the Apple private key")
		}
		data, err := os.ReadFile(r.privateKeyFile)
		if err != nil {
			return nil, errors.New("failed to load the Apple private key")
		}
		pemText = string(data)
	} else {
		pemText = strings.ReplaceAll(pemText, `\n`, "\n")
	}
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(pemText))
	if err != nil {
		return nil, errors.New("failed to parse the Apple private key")
	}
	r.key = key
	return key, nil
}

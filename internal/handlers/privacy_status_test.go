package handlers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrivacyStatusRequiresExplicitAIOptInAndNeverLeaksKeys(t *testing.T) {
	old := configs.Env
	t.Cleanup(func() { configs.Env = old })
	for _, tt := range []struct {
		key     string
		optIn   string
		enabled bool
	}{{"", "true", false}, {"secret-openai-key", "false", false}, {"secret-openai-key", "true", true}} {
		configs.Env = configs.Config{OpenaiKey: tt.key, ImageDescription: configs.ImageDescriptionEnv{Enabled: tt.optIn}}
		app := fiber.New()
		h := newNuboStatusHandler(nil, "test")
		app.Get("/privacy/status", h.PrivacyStatusHandler)
		response, err := app.Test(httptest.NewRequest("GET", "/privacy/status", nil))
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(response.Body)
		response.Body.Close()
		var body struct {
			Enabled   bool     `json:"imageDescriptionEnabled"`
			Providers []string `json:"providers"`
		}
		if err := json.Unmarshal(data, &body); err != nil {
			t.Fatal(err)
		}
		if body.Enabled != tt.enabled || (len(body.Providers) > 0) != tt.enabled {
			t.Fatalf("wrong state: %s", data)
		}
		if strings.Contains(string(data), "secret-openai-key") {
			t.Fatal("key exposed")
		}
	}
}

func TestPrivacyStatusIncludesConfiguredDomesticLoginProviders(t *testing.T) {
	old := configs.Env
	t.Cleanup(func() { configs.Env = old })
	configs.Env = configs.Config{OAuthNaverID: "private-naver-client", OAuthKakaoID: "private-kakao-client"}
	app := fiber.New()
	h := newNuboStatusHandler(nil, "test")
	app.Get("/privacy/status", h.PrivacyStatusHandler)
	response, err := app.Test(httptest.NewRequest("GET", "/privacy/status", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	text := string(data)
	if !strings.Contains(text, "Naver") || !strings.Contains(text, "Kakao") || strings.Contains(text, "private-") {
		t.Fatalf("wrong provider disclosure: %s", data)
	}
}

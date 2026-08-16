package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/resend/resend-go/v2"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

func TestResendMailerSendsProviderNeutralMessage(t *testing.T) {
	var request resend.SendEmailRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			t.Fatalf("path = %q, want /emails", r.URL.Path)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "signup/42" {
			t.Fatalf("idempotency key = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	client := resend.NewCustomClient(server.Client(), "re_test")
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = baseURL
	mailer := &ResendMailer{
		client:    client,
		apiKey:    "re_test",
		fromEmail: "noreply@example.com",
		fromName:  "NUBO",
	}

	delivery, err := mailer.Send(models.MailMessage{
		To:             "member@example.com",
		Subject:        "Verify",
		HTML:           "<p>code</p>",
		Text:           "code",
		IdempotencyKey: "signup/42",
		Tags:           map[string]string{"type": "signup"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if delivery.Provider != "resend" || delivery.MessageID != "email_123" {
		t.Fatalf("delivery = %#v", delivery)
	}
	if request.From != `"NUBO" <noreply@example.com>` || request.To[0] != "member@example.com" {
		t.Fatalf("request addresses = from %q, to %#v", request.From, request.To)
	}
	if request.Text != "code" || request.Html != "<p>code</p>" {
		t.Fatalf("request bodies were not preserved")
	}
}

func TestResendMailerRequiresKeyAndValidSender(t *testing.T) {
	tests := []ResendMailer{
		{apiKey: "", fromEmail: "noreply@example.com"},
		{apiKey: "re_test", fromEmail: ""},
		{apiKey: "re_test", fromEmail: "not-an-email"},
	}
	for _, mailer := range tests {
		if mailer.Configured() {
			t.Fatalf("mailer unexpectedly configured: %#v", mailer)
		}
	}
}

func TestResendMailerStatusReportsDomainVerification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains" {
			t.Fatalf("path = %q, want /domains", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"name":"example.com","status":"verified"}]}`))
	}))
	defer server.Close()

	client := resend.NewCustomClient(server.Client(), "re_test")
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = baseURL
	mailer := &ResendMailer{
		client:    client,
		apiKey:    "re_test",
		fromEmail: "noreply@example.com",
		fromName:  "NUBO",
	}
	status := mailer.Status()
	if !status.Configured || status.DomainStatus != "verified" {
		t.Fatalf("status = %#v", status)
	}
}

func TestDefaultResendFromEmailUsesConfiguredSiteDomain(t *testing.T) {
	if got := defaultResendFromEmail("https://community.example.com/path"); got != "noreply@community.example.com" {
		t.Fatalf("default sender = %q", got)
	}
	if got := defaultResendFromEmail("http://localhost:3000"); got != "" {
		t.Fatalf("localhost sender = %q", got)
	}
}

func TestResendLiveDelivery(t *testing.T) {
	if os.Getenv("NUBO_RUN_MAIL_INTEGRATION") != "1" {
		t.Skip("set NUBO_RUN_MAIL_INTEGRATION=1 to send to Resend's delivered test sink")
	}
	previous := configs.Env
	configs.Env = configs.Config{
		Domain:          os.Getenv("GOAPI_DOMAIN"),
		Title:           os.Getenv("GOAPI_TITLE"),
		ResendKey:       os.Getenv("RESEND_API_KEY"),
		ResendFromEmail: os.Getenv("RESEND_FROM_EMAIL"),
		ResendFromName:  os.Getenv("RESEND_FROM_NAME"),
	}
	t.Cleanup(func() { configs.Env = previous })

	mailer := NewResendMailer()
	if !mailer.Configured() {
		t.Fatal("live Resend configuration is incomplete")
	}
	delivery, err := mailer.Send(models.MailMessage{
		To:             "delivered+transactional-stabilization@resend.dev",
		Subject:        "NUBO transactional mail integration test",
		HTML:           "<p>NUBO transactional email delivery test.</p>",
		Text:           "NUBO transactional email delivery test.",
		IdempotencyKey: "integration/transactional-mail/" + time.Now().UTC().Format("2006-01-02"),
		Tags:           map[string]string{"type": "integration-test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Resend accepted integration message %s", delivery.MessageID)
}

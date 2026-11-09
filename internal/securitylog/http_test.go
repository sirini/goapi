package securitylog_test

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/securitylog"
	"github.com/sirini/goapi/pkg/utils"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrivateLogRoutesAndExportAudit(t *testing.T) {
	previous := securitylog.Default
	oldEnv := configs.Env
	t.Cleanup(func() { securitylog.Default = previous; configs.Env = oldEnv })
	store, err := securitylog.New(t.TempDir(), 7, 1024*1024, true)
	if err != nil {
		t.Fatal(err)
	}
	securitylog.Default = store
	configs.Env.JWTSecretKey = "test-secret"
	app := fiber.New()
	app.Use(securitylog.Middleware)
	app.Get("/admin/system/logs", middlewares.AdminMiddleware(func(uint) bool { return true }), securitylog.ReadHandler)
	app.Post("/post/:id", func(c fiber.Ctx) error { return c.SendStatus(204) })
	app.Get("/public", func(c fiber.Ctx) error { return c.SendStatus(200) })
	for _, uid := range []uint{0, 2, 1} {
		request := httptest.NewRequest("GET", "/admin/system/logs?format=csv", nil)
		if uid > 0 {
			token, _ := utils.GenerateAccessToken(uid, 1)
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if uid == 1 {
			if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/csv") || !strings.Contains(string(body), "time_utc") {
				t.Fatal("admin export failed")
			}
		} else if strings.Contains(string(body), "time_utc") {
			t.Fatal("unauthorized CSV disclosure")
		}
	}
	req := httptest.NewRequest("POST", "/post/private-name?email=secret@example.org", strings.NewReader("password-secret"))
	req.Header.Set("X-Forwarded-For", "203.0.113.88")
	response, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	response, err = app.Test(httptest.NewRequest("GET", "/public", nil))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	data, err := store.Export(time.Now().UTC().Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, secret := range []string{"private-name", "secret@example.org", "password-secret", "203.0.113.88", "/public"} {
		if strings.Contains(text, secret) {
			t.Fatalf("unexpected sensitive data %s", secret)
		}
	}
	if !strings.Contains(text, "/post/:id") || !strings.Contains(text, "#csv-export") {
		t.Fatal("missing template or export audit")
	}
}

func TestAuditRecordsSubjectsAndBlocksAccessWhenIntegrityFails(t *testing.T) {
	previous := securitylog.Audit
	oldEnv := configs.Env
	t.Cleanup(func() { securitylog.Audit = previous; configs.Env = oldEnv })
	dir := t.TempDir()
	store, err := securitylog.NewAudit(dir, 732)
	if err != nil {
		t.Fatal(err)
	}
	securitylog.Audit = store
	configs.Env.JWTSecretKey = "audit-test-secret"
	token, _ := utils.GenerateAccessToken(1, 1)
	ran := 0
	app := fiber.New()
	group := app.Group("/admin", middlewares.AdminMiddleware(func(uint) bool { return true }), securitylog.AuditMiddleware)
	group.Get("/user/list", func(c fiber.Ctx) error {
		ran++
		return c.JSON(fiber.Map{"result": fiber.Map{"item": []fiber.Map{{"uid": 7, "id": "private@example.org"}, {"uid": 8}}}})
	})
	request := func() {
		req := httptest.NewRequest("GET", "/admin/user/list", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}
	request()
	if ran != 1 {
		t.Fatal("valid request blocked")
	}
	day := time.Now().UTC().Format("2006-01-02")
	page, err := store.Read(day, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[0].Actor != 1 || len(page.Items[0].Subjects) != 2 {
		t.Fatalf("missing subject audit: %+v", page)
	}
	data, err := store.Export(day)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private@example.org") {
		t.Fatal("email copied into audit")
	}
	file := filepath.Join(dir, day+".jsonl")
	if err := os.WriteFile(file, []byte("corrupt\n"), 0600); err != nil {
		t.Fatal(err)
	}
	request()
	if ran != 1 {
		t.Fatal("admin action ran despite corrupt audit journal")
	}
}

func TestOnlyExplicitTrustedProxyCanSupplyRealIP(t *testing.T) {
	old := configs.Env
	t.Cleanup(func() { configs.Env = old })
	app := fiber.New()
	app.Get("/peer", func(c fiber.Ctx) error { return c.SendString(securitylog.RequestPeer(c)) })
	read := func() string {
		req := httptest.NewRequest("GET", "/peer", nil)
		req.Header.Set("X-Real-IP", "198.51.100.41")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		data, _ := io.ReadAll(res.Body)
		return string(data)
	}
	configs.Env.TrustedProxyCIDRs = ""
	peer := read()
	if peer == "198.51.100.41" {
		t.Fatal("untrusted forwarded address accepted")
	}
	configs.Env.TrustedProxyCIDRs = peer + "/32"
	if got := read(); got != "198.51.100.41" {
		t.Fatalf("trusted proxy not used: %s (peer=%s)", got, peer)
	}
	configs.Env.TrustedProxyCIDRs = "0.0.0.0/0"
	if got := read(); got != peer {
		t.Fatal("world trust accepted")
	}
}

func TestAuditReviewAndCSVRequirePurpose(t *testing.T) {
	old := securitylog.Audit
	t.Cleanup(func() { securitylog.Audit = old })
	store, err := securitylog.NewAudit(t.TempDir(), 732)
	if err != nil {
		t.Fatal(err)
	}
	securitylog.Audit = store
	app := fiber.New()
	group := app.Group("/admin", func(c fiber.Ctx) error { c.Locals("securityLogActor", uint(1)); return c.Next() }, securitylog.AuditMiddleware)
	group.Get("/system/audit", securitylog.AuditReadHandler)
	group.Post("/system/audit/review", securitylog.AuditReviewHandler)
	day := time.Now().UTC().Format("2006-01-02")
	month := time.Now().UTC().Format("2006-01")
	for _, tt := range []struct {
		method, path string
		status       int
	}{{"GET", "/admin/system/audit?format=csv&date=" + day, 400}, {"GET", "/admin/system/audit?format=csv&purpose=privacy-review&date=" + day, 200}, {"POST", "/admin/system/audit/review?month=" + month, 200}, {"POST", "/admin/system/audit/review?month=2999-01", 400}} {
		res, err := app.Test(httptest.NewRequest(tt.method, tt.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tt.status {
			t.Fatalf("%s got %d want %d", tt.path, res.StatusCode, tt.status)
		}
	}
	page, err := store.Read(day, 0)
	if err != nil {
		t.Fatal(err)
	}
	review, purpose := false, false
	for _, entry := range page.Items {
		review = review || entry.Phase == "review-confirmed"
		purpose = purpose || entry.Purpose == "privacy-review"
	}
	if !review || !purpose {
		t.Fatal("missing review or export purpose audit")
	}
}

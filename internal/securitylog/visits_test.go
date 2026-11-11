package securitylog_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/securitylog"
	"github.com/sirini/goapi/pkg/utils"
)

func TestVisitorIdentityPagesAndAdminAudit(t *testing.T) {
	previous, audit, env := securitylog.Visits, securitylog.Audit, configs.Env
	t.Cleanup(func() { securitylog.Visits = previous; securitylog.Audit = audit; configs.Env = env })
	configs.Env.JWTSecretKey = "visitor-test-secret"
	store, err := securitylog.New(t.TempDir(), 90, 1024*1024, false)
	if err != nil {
		t.Fatal(err)
	}
	store.EnableFullIP()
	securitylog.Visits = store
	securitylog.Audit, err = securitylog.NewAudit(t.TempDir(), 732)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	active := func(uid uint) bool { return uid != 3 }
	app.Post("/visit", securitylog.VisitHandler(active))
	app.Get("/admin/visits", middlewares.AdminMiddleware(active), securitylog.VisitsReadHandler)
	app.Get("/admin/stats", middlewares.AdminMiddleware(active), securitylog.VisitStatisticsHandler)
	for _, tc := range []struct {
		uid    uint
		path   string
		status int
	}{
		{0, "/", 200}, {2, "/board/12", 200}, {1, "/admin/members", 200}, {3, "/", 401}, {0, "/search?password=secret", 400}, {0, "//evil", 400},
	} {
		req := httptest.NewRequest("POST", "/visit?userUid=1", strings.NewReader(`{"path":"`+tc.path+`","userUid":1}`))
		req.Header.Set("Content-Type", "application/json")
		if tc.uid > 0 {
			token, _ := utils.GenerateAccessToken(tc.uid, 1)
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != tc.status {
			t.Fatalf("uid %d path %s: %d", tc.uid, tc.path, resp.StatusCode)
		}
	}
	day := time.Now().UTC().Format("2006-01-02")
	page, err := store.Read(day, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 3 || !page.FullIP {
		t.Fatalf("bad page: %+v", page)
	}
	actors := map[uint]bool{}
	for _, entry := range page.Items {
		actors[entry.Actor] = true
	}
	if !actors[0] || !actors[1] || !actors[2] {
		t.Fatal("caller identity was trusted")
	}
	page, err = securitylog.Audit.Read(day, 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Route != "/admin/members" || page.Items[0].Phase != "page-visit" {
		t.Fatalf("missing admin navigation: %+v %v", page, err)
	}
	for _, uid := range []uint{0, 2, 1} {
		for _, path := range []string{"/admin/visits?format=csv", "/admin/stats?unit=day"} {
			req := httptest.NewRequest("GET", path, nil)
			if uid > 0 {
				token, _ := utils.GenerateAccessToken(uid, 1)
				req.Header.Set("Authorization", "Bearer "+token)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if uid != 1 && (strings.Contains(string(body), "time_utc") || strings.Contains(string(body), "counts")) {
				t.Fatal("log disclosure")
			}
			if uid == 1 && strings.Contains(path, "stats") {
				var stats struct{ Counts map[string]int }
				if json.Unmarshal(body, &stats) != nil || stats.Counts[day] != 3 {
					t.Fatalf("bad counts %s", body)
				}
			}
			if uid == 1 && strings.Contains(path, "visits") && !strings.HasPrefix(string(body), "\xef\xbb\xbf") {
				t.Fatal("missing CSV BOM")
			}
		}
	}
	// Counts must include all pages beyond the 100-row UI page and invalidate cached days after writes.
	for range 101 {
		if err := store.Append(securitylog.Entry{Method: "PAGE", Route: "/", Peer: "192.0.2.10"}); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		req := httptest.NewRequest("GET", "/admin/stats?unit=day", nil)
		token, _ := utils.GenerateAccessToken(1, 1)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var stats struct{ Counts map[string]int }
		if json.Unmarshal(body, &stats) != nil || stats.Counts[day] != 104 {
			t.Fatalf("stale or paginated count: %s", body)
		}
	}
	csv, err := store.Export(day)
	if err != nil || !strings.Contains(string(csv), "192.0.2.10") {
		t.Fatal("full visitor IP missing from CSV")
	}

}

func TestVisitSensitivePathRedaction(t *testing.T) {
	old, env := securitylog.Visits, configs.Env
	t.Cleanup(func() { securitylog.Visits = old; configs.Env = env })
	store, err := securitylog.New(t.TempDir(), 7, 1024*1024, false)
	if err != nil {
		t.Fatal(err)
	}
	securitylog.Visits = store
	app := fiber.New()
	app.Post("/visit", securitylog.VisitHandler(nil))
	for _, path := range []string{"/search/0/private-search", "/board/news/search/0/private-search/1", "/auth/change-password/2/private-code", "/community/auth/change-password/2/private-code", "/community/%73earch/0/private-search", "/AUTH/CHANGE-PASSWORD/2/private-code", "/SEARCH/0/private-search"} {
		req := httptest.NewRequest("POST", "/visit", strings.NewReader(`{"path":"`+path+`"}`))
		req.Header.Set("Content-Type", "application/json")
		response, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatal(response.StatusCode)
		}
	}
	data, err := store.Export(time.Now().UTC().Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-search") || strings.Contains(string(data), "private-code") {
		t.Fatal("sensitive path retained")
	}
}

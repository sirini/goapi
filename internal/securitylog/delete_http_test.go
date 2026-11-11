package securitylog_test

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/securitylog"
	"github.com/sirini/goapi/pkg/utils"
)

func TestManualDeletionAuthorizationConfirmationAndEvidence(t *testing.T) {
	oldVisits, oldAudit, env := securitylog.Visits, securitylog.Audit, configs.Env
	t.Cleanup(func() { securitylog.Visits = oldVisits; securitylog.Audit = oldAudit; configs.Env = env })
	configs.Env.JWTSecretKey = "manual-deletion-test-secret"
	dir, auditDir := t.TempDir(), t.TempDir()
	var err error
	securitylog.Visits, err = securitylog.NewVisits(dir)
	if err != nil {
		t.Fatal(err)
	}
	securitylog.Audit, err = securitylog.NewAudit(auditDir, 732)
	if err != nil {
		t.Fatal(err)
	}
	day := time.Now().UTC().AddDate(-4, 0, 0).Format("2006-01-02")
	path := filepath.Join(dir, day+".jsonl")
	if err := os.WriteFile(path, []byte(`{"time":"`+day+`T00:00:00Z","method":"PAGE","route":"/","actor":0}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	admin := app.Group("/admin", middlewares.AdminMiddleware(func(uid uint) bool { return uid != 3 }), securitylog.AuditMiddleware)
	admin.Delete("/system/records/:kind", securitylog.DeleteRecordsHandler)
	admin.Get("/stats", securitylog.VisitStatisticsHandler)
	request := func(uid uint, kind, date, confirm, reason string) (int, string) {
		req := httptest.NewRequest("DELETE", "/admin/system/records/"+kind+"?date="+date, strings.NewReader(`{"confirmDate":"`+confirm+`","reason":"`+reason+`"}`))
		req.Header.Set("Content-Type", "application/json")
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
		return resp.StatusCode, string(body)
	}
	for _, uid := range []uint{0, 2, 3} {
		request(uid, "visits", day, day, "retention-policy")
		if _, err := os.Stat(path); err != nil {
			t.Fatal("unauthorized delete")
		}
	}
	for _, tc := range []struct{ kind, date, confirm, reason string }{
		{"visits", day, "wrong", "retention-policy"}, {"visits", day, day, "invalid"}, {"logs", day, day, "retention-policy"},
		{"audit", time.Now().UTC().Format("2006-01-02"), time.Now().UTC().Format("2006-01-02"), "retention-policy"},
	} {
		status, _ := request(1, tc.kind, tc.date, tc.confirm, tc.reason)
		if status != 400 {
			t.Fatalf("invalid deletion status %d", status)
		}
	}
	// Historical statistics must still include records older than three years.
	req := httptest.NewRequest("GET", "/admin/stats?year="+day[:4]+"&unit=day", nil)
	token, _ := utils.GenerateAccessToken(1, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `"`+day+`":1`) {
		t.Fatalf("historical stats excluded: %s", body)
	}
	status, result := request(1, "visits", day, day, "retention-policy")
	if status != 200 || !strings.Contains(result, `"deleted":true`) {
		t.Fatalf("delete failed: %d %s", status, result)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("data retained after explicit delete")
	}
	page, err := securitylog.Audit.Read(time.Now().UTC().Format("2006-01-02"), 0)
	if err != nil {
		t.Fatal(err)
	}
	phases := map[string]bool{}
	for _, entry := range page.Items {
		if entry.Reference == day && entry.Purpose == "retention-policy" && entry.Actor == 1 && entry.Route == "/admin/system/records/visits" {
			phases[entry.Phase] = true
		}
	}
	if !phases["delete-intent"] || !phases["delete-completed"] {
		t.Fatal("missing deletion evidence")
	}
	// A broken audit must stop deletion before files are touched.
	if err := os.WriteFile(path, []byte("still retained\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(auditDir, time.Now().UTC().Format("2006-01-02")+".jsonl"), []byte("tampered\n"), 0600); err != nil {
		t.Fatal(err)
	}
	status, _ = request(1, "visits", day, day, "test-cleanup")
	if status != 503 {
		t.Fatalf("audit failure not closed: %d", status)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("unaudited deletion allowed")
	}
}

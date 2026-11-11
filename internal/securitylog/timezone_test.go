package securitylog

import (
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestKSTDayReadCSVAndCountsSpanUTCBoundariesWithoutRewriting(t *testing.T) {
	for _, kind := range []string{"visits", "audit"} {
		t.Run(kind, func(t *testing.T) {
			var store *Store
			var err error
			if kind == "audit" {
				store, err = NewAudit(t.TempDir(), 0)
			} else {
				store, err = NewVisits(t.TempDir())
			}
			if err != nil {
				t.Fatal(err)
			}
			for i, value := range []string{"2025-12-31T14:59:59Z", "2025-12-31T15:00:00Z", "2026-01-01T14:59:59.999Z", "2026-01-01T15:00:00Z"} {
				stamp, err := time.Parse(time.RFC3339Nano, value)
				if err != nil {
					t.Fatal(err)
				}
				store.now = func() time.Time { return stamp }
				if err := store.Append(Entry{Route: []string{"/excluded-before", "/included-start", "/included-end", "/excluded-after"}[i]}); err != nil {
					t.Fatal(err)
				}
			}
			before := map[string]string{}
			for _, day := range []string{"2025-12-31", "2026-01-01"} {
				data, err := os.ReadFile(store.dayPath(day))
				if err != nil {
					t.Fatal(err)
				}
				before[day] = string(data)
			}
			page, err := store.ReadInLocation("2026-01-01", 0, KST)
			if err != nil || len(page.Items) != 2 || page.Timezone != "Asia/Seoul" {
				t.Fatalf("KST window: %+v %v", page, err)
			}
			if page.Items[0].Route != "/included-end" || page.Items[1].Route != "/included-start" {
				t.Fatal("wrong day membership or sort")
			}
			csv, err := store.ExportInLocation("2026-01-01", KST)
			if err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{"time_kst", "2026-01-01 00:00:00", "2026-01-01 23:59:59"} {
				if !strings.Contains(string(csv), text) {
					t.Fatalf("missing %s: %s", text, csv)
				}
			}
			if strings.Contains(string(csv), "excluded-") {
				t.Fatal("export contains another KST day")
			}
			for _, day := range []string{"2025-12-31", "2026-01-01"} {
				data, _ := os.ReadFile(store.dayPath(day))
				if string(data) != before[day] {
					t.Fatal("signed UTC history rewritten")
				}
			}
			if kind == "visits" {
				for range 2 {
					count, err := store.countDayInLocation("2026-01-01", KST)
					if err != nil || count != 2 {
						t.Fatalf("KST count %d %v", count, err)
					}
				}
				old := Visits
				Visits = store
				t.Cleanup(func() { Visits = old })
				store.now = func() time.Time { return time.Date(2026, 1, 2, 12, 0, 0, 0, KST) }
				app := fiber.New()
				app.Get("/stats", VisitStatisticsHandler)
				for _, unit := range []string{"month", "year"} {
					resp, err := app.Test(httptest.NewRequest("GET", "/stats?year=2026&timezone=Asia%2FSeoul&unit="+unit, nil))
					if err != nil {
						t.Fatal(err)
					}
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					key := "2026-01"
					if unit == "year" {
						key = "2026"
					}
					if !strings.Contains(string(body), `"`+key+`":3`) {
						t.Fatalf("year boundary total: %s", body)
					}
				}
			} else {
				if err := os.WriteFile(store.dayPath("2025-12-31"), []byte(strings.Replace(before["2025-12-31"], "/excluded-before", "/tampered", 1)), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := store.ExportInLocation("2026-01-01", KST); err == nil {
					t.Fatal("tampered adjacent UTC file was ignored")
				}
			}
		})
	}
}

func TestKSTReadHandlerAndMonthReviewIncludePreviousUTCDate(t *testing.T) {
	oldAudit := Audit
	t.Cleanup(func() { Audit = oldAudit })
	var err error
	Audit, err = NewAudit(t.TempDir(), 0)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2025, 12, 31, 15, 0, 0, 0, time.UTC)
	Audit.now = func() time.Time { return stamp }
	if err := Audit.Append(Entry{Route: "/admin"}); err != nil {
		t.Fatal(err)
	}
	Audit.now = time.Now
	app := fiber.New()
	app.Get("/audit", AuditReadHandler)
	app.Post("/review", AuditReviewHandler)
	resp, err := app.Test(httptest.NewRequest("GET", "/audit?date=2026-01-01&format=csv&timezone=Asia%2FSeoul", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || resp.Header.Get("X-Nubo-Log-Timezone") != "Asia/Seoul" || !strings.Contains(string(body), "2026-01-01 00:00:00") {
		t.Fatalf("KST HTTP export: %d %s", resp.StatusCode, body)
	}
	path := Audit.dayPath("2025-12-31")
	data, _ := os.ReadFile(path)
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), "/admin", "/tamper", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	resp, err = app.Test(httptest.NewRequest("POST", "/review?month=2026-01&timezone=Asia%2FSeoul", nil))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 503 {
		t.Fatalf("KST month missed previous UTC date: %d", resp.StatusCode)
	}
	resp, err = app.Test(httptest.NewRequest("GET", "/audit?timezone=invalid", nil))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatal("unknown timezone accepted")
	}
}

func TestLegacyPendingDeletionSealIsVerifiedWithoutMutation(t *testing.T) {
	s, err := NewAudit(t.TempDir(), 0)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 1, 1, 15, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return stamp }
	if err := s.Append(Entry{Route: "/legacy"}); err != nil {
		t.Fatal(err)
	}
	path := s.dayPath("2026-01-01")
	before, _ := os.ReadFile(path)
	if err := os.Rename(path+".seal", path+".seal.deleting"); err != nil {
		t.Fatal(err)
	}
	page, err := s.ReadInLocation("2026-01-02", 0, KST)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("cannot read legacy pending journal: %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("legacy data rewritten")
	}
	if _, err := os.Stat(path + ".seal.deleting"); err != nil {
		t.Fatal("pending seal deleted")
	}
	if _, err := os.Stat(path + ".seal"); !os.IsNotExist(err) {
		t.Fatal("read modified the seal path")
	}
}

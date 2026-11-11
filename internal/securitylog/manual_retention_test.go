package securitylog

import (
	"strings"
	"testing"
	"time"
)

func TestLongTermRecordsSurviveStartupPruningReadsAndWrites(t *testing.T) {
	for _, kind := range []string{"visits", "audit"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			open := func() (*Store, error) {
				if kind == "audit" {
					return NewAudit(dir, 732)
				}
				return NewVisits(dir)
			}
			s, err := open()
			if err != nil {
				t.Fatal(err)
			}
			past := time.Now().UTC().AddDate(-4, 0, 0)
			day := past.Format("2006-01-02")
			s.now = func() time.Time { return past }
			if err := s.Append(Entry{Method: "PAGE", Route: "/admin", Actor: 1, Peer: "192.0.2.10"}); err != nil {
				t.Fatal(err)
			}
			s, err = open()
			if err != nil {
				t.Fatal(err)
			} // startup must not prune four-year-old records
			if s.days != 0 {
				t.Fatal("automatic expiry remains enabled")
			}
			if err := s.Prune(); err != nil {
				t.Fatal(err)
			}
			if err := s.Append(Entry{Route: "/today"}); err != nil {
				t.Fatal(err)
			}
			page, err := s.Read(day, 0)
			if err != nil || len(page.Items) != 1 || !page.ManualDeletionOnly {
				t.Fatalf("history lost: %+v %v", page, err)
			}
			csv, err := s.Export(day)
			if err != nil || !strings.Contains(string(csv), "192.0.2.10") {
				t.Fatalf("old export lost: %v", err)
			}
			page, err = s.Read(time.Now().UTC().Format("2006-01-02"), 0)
			if err != nil || len(page.Items) != 1 {
				t.Fatal("unselected day changed")
			}
			s, err = open()
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Append(Entry{Route: "/after-restart"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

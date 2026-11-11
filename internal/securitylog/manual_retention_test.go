package securitylog

import (
	"os"
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
			if _, err := s.DeleteDay(time.Now().UTC().Format("2006-01-02")); err == nil {
				t.Fatal("live day deletion allowed")
			}
			deleted, err := s.DeleteDay(day)
			if err != nil || !deleted {
				t.Fatalf("manual deletion failed: %v", err)
			}
			page, err = s.Read(day, 0)
			if err != nil || len(page.Items) != 0 {
				t.Fatal("deleted data remains")
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

func TestAuditDeletionRetriesRecoverInterruptedSealMoves(t *testing.T) {
	for _, dataRemoved := range []bool{false, true} {
		t.Run(map[bool]string{false: "before-unlink", true: "after-unlink"}[dataRemoved], func(t *testing.T) {
			s, err := NewAudit(t.TempDir(), 0)
			if err != nil {
				t.Fatal(err)
			}
			past := time.Now().UTC().AddDate(-4, 0, 0)
			s.now = func() time.Time { return past }
			if err := s.Append(Entry{Route: "/admin"}); err != nil {
				t.Fatal(err)
			}
			day := past.Format("2006-01-02")
			path := s.dayPath(day)
			s.now = time.Now
			if err := os.Rename(path+".seal", path+".seal.deleting"); err != nil {
				t.Fatal(err)
			}
			if dataRemoved {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			deleted, err := s.DeleteDay(day)
			if err != nil || !deleted {
				t.Fatalf("cannot resume deletion: %v", err)
			}
			for _, suffix := range []string{"", ".seal", ".seal.deleting"} {
				if _, err := os.Lstat(path + suffix); !os.IsNotExist(err) {
					t.Fatalf("artifact remains: %s", suffix)
				}
			}
		})
	}
}

func TestAuditDeletionRefusesCorruptionAndInvalidDates(t *testing.T) {
	s, err := NewAudit(t.TempDir(), 0)
	if err != nil {
		t.Fatal(err)
	}
	past := time.Now().UTC().AddDate(-4, 0, 0)
	s.now = func() time.Time { return past }
	day := past.Format("2006-01-02")
	if err := s.Append(Entry{Route: "/admin"}); err != nil {
		t.Fatal(err)
	}
	s.now = time.Now
	path := s.dayPath(day)
	if err := os.WriteFile(path, []byte("corrupt\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, date := range []string{day, "../escape", "2026-02-30", time.Now().UTC().Format("2006-01-02"), "9999-12-31"} {
		if _, err := s.DeleteDay(date); err == nil {
			t.Fatalf("deletion allowed: %s", date)
		}
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("corrupt file was silently erased")
	}
}

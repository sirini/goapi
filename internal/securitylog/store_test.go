package securitylog

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRetentionMaskingAndCSV(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir, 7, 1024, true)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	old := filepath.Join(dir, "2026-11-03.jsonl")
	if err := os.WriteFile(old, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(Entry{Method: "POST", Route: "/board/:id", Peer: "192.0.2.123", Actor: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("expired data not removed")
	}
	page, err := s.Read("2026-11-10", 0)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	if page.Items[0].Peer != "192.0.2.0/24" {
		t.Fatal(page.Items[0])
	}
	if err := s.Append(Entry{Method: "=HYPERLINK(\"bad\")", Route: "+formula"}); err != nil {
		t.Fatal(err)
	}
	data, err := s.Export("2026-11-10")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff"))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[2][1][0] != '\'' || rows[2][2][0] != '\'' {
		t.Fatalf("unsafe CSV: %q", data)
	}
	if _, err := s.Export("../../etc/passwd"); err == nil {
		t.Fatal("accepted traversal")
	}
	now = now.AddDate(0, 0, 7)
	if err := s.Prune(); err != nil {
		t.Fatal(err)
	}
	page, err = s.Read("2026-11-10", 0)
	if err != nil || len(page.Items) != 0 {
		t.Fatal("expired day retained")
	}
}

func TestBoundedStorageAndIPOptOut(t *testing.T) {
	s, err := New(t.TempDir(), 7, 1024, false)
	if err != nil {
		t.Fatal(err)
	}
	for range 100 {
		if err := s.Append(Entry{Method: "POST", Route: "/auth/signin", Peer: "192.0.2.1"}); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.Read(time.Now().UTC().Format("2006-01-02"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if page.Dropped == 0 || len(page.Items) == 0 {
		t.Fatalf("missing cap: %+v", page)
	}
	for _, e := range page.Items {
		if e.Peer != "" {
			t.Fatal("IP opt-out ignored")
		}
	}
	if got := MaskPeer("2001:db8:1234:abcd::1"); got != "2001:db8:1234::/48" {
		t.Fatal(got)
	}
}

func TestPrivateDirectoryRejectsPublicUploadsAndSymlinks(t *testing.T) {
	root := t.TempDir()
	upload := filepath.Join(root, "upload")
	if err := os.Mkdir(upload, 0700); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{upload, filepath.Join(upload, "logs")} {
		if ValidatePrivateDirectory(dir, upload) == nil {
			t.Fatal("accepted public log directory")
		}
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(upload, alias); err != nil {
		t.Fatal(err)
	}
	if ValidatePrivateDirectory(filepath.Join(alias, "logs"), upload) == nil {
		t.Fatal("accepted symlink into public uploads")
	}
	if err := ValidatePrivateDirectory(filepath.Join(root, "private"), upload); err != nil {
		t.Fatal(err)
	}
}

func TestAuditIntegrityDetectsModificationTruncationAndMissingData(t *testing.T) {
	dir := t.TempDir()
	s, err := NewAudit(dir, 732)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	for range 2 {
		if err := s.Append(Entry{Method: "GET", Route: "/admin/user/load", Peer: "192.0.2.123", Actor: 1, Subjects: []uint{7}}); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.Read("2026-11-10", 0)
	if err != nil || len(page.Items) != 2 || !page.IntegrityChecked || page.Items[0].Peer != "192.0.2.123" {
		t.Fatalf("page %+v err %v", page, err)
	}
	path := s.dayPath("2026-11-10")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	altered := strings.Replace(string(original), "192.0.2.123", "192.0.2.124", 1)
	if err := os.WriteFile(path, []byte(altered), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Export("2026-11-10"); err == nil {
		t.Fatal("tampered export allowed")
	}
	lines := strings.SplitAfter(string(original), "\n")
	if err := os.WriteFile(path, []byte(lines[0]), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read("2026-11-10", 0); err == nil {
		t.Fatal("truncated journal allowed")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Export("2026-11-10"); err == nil {
		t.Fatal("missing journal ignored")
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Export("2026-11-10"); err != nil {
		t.Fatal(err)
	}
	now = now.AddDate(0, 0, 733)
	if err := s.Prune(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".seal"); err != nil {
		t.Fatal("audit seal must remain until explicit deletion")
	}
}

func TestAuditStartupDoesNotEraseOldDataToRecoverMissingKey(t *testing.T) {
	dir := t.TempDir()
	for _, suffix := range []string{".jsonl", ".jsonl.seal", ".jsonl.seal.tmp"} {
		if err := os.WriteFile(filepath.Join(dir, "2020-01-01"+suffix), []byte("expired"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewAudit(dir, 732); err == nil {
		t.Fatal("missing key must fail without deleting any old files")
	}
	files, err := filepath.Glob(filepath.Join(dir, "2020-01-01*"))
	if err != nil || len(files) != 3 {
		t.Fatalf("old metadata must remain: %v %v", files, err)
	}
}

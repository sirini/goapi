// Package securitylog stores bounded operational metadata, never request payloads.
package securitylog

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	Reference    string `json:"reference,omitempty"`
	PeerSource   string `json:"peerSource,omitempty"`
	RequestID    string `json:"requestId,omitempty"`
	ResultCode   *int   `json:"resultCode,omitempty"`
	Purpose      string `json:"purpose,omitempty"`
	Subjects     []uint `json:"subjects,omitempty"`
	Phase        string `json:"phase,omitempty"`
	PreviousHash string `json:"previousHash,omitempty"`
	Hash         string `json:"hash,omitempty"`
	Time         string `json:"time"`
	Method       string `json:"method"`
	Route        string `json:"route"`
	Status       int    `json:"status"`
	DurationMS   int64  `json:"durationMs"`
	Actor        uint   `json:"actor"`
	Peer         string `json:"peer,omitempty"`
}

type Store struct {
	integrityKey []byte
	fullIP       bool
	strict       bool

	mu        sync.Mutex
	dir       string
	days      int
	maxBytes  int64
	maskedIP  bool
	now       func() time.Time
	lastError bool
	dropped   uint64
}

type Page struct {
	FullIP           bool `json:"fullIP"`
	IntegrityChecked bool `json:"integrityChecked"`

	Items         []Entry `json:"items"`
	Next          int     `json:"next"`
	RetentionDays int     `json:"retentionDays"`
	MaskedIP      bool    `json:"maskedIP"`
	WriteError    bool    `json:"writeError"`
	Dropped       uint64  `json:"dropped"`
}

func New(dir string, days int, maxBytes int64, maskedIP bool) (*Store, error) {
	if days < 1 || days > 800 || maxBytes < 1024 || maxBytes > 100*1024*1024 {
		return nil, errors.New("invalid security log limits")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, days: days, maxBytes: maxBytes, maskedIP: maskedIP, now: time.Now}
	return s, s.Prune()
}

// Peer is the network peer, not an untrusted X-Forwarded-For value.
func MaskPeer(value string) string {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return ""
	}
	ip = ip.Unmap()
	bits := 24
	if ip.Is6() {
		bits = 48
	}
	return netip.PrefixFrom(ip, bits).Masked().String()
}

func (s *Store) dayPath(day string) string { return filepath.Join(s.dir, day+".jsonl") }
func (s *Store) pruneLocked() error {
	files, err := filepath.Glob(filepath.Join(s.dir, "*.jsonl*"))
	if err != nil {
		return err
	}
	cutoff := s.now().UTC().AddDate(0, 0, -s.days+1).Format("2006-01-02")
	for _, path := range files {
		name := filepath.Base(path)
		if !strings.HasSuffix(name, ".jsonl") && !strings.HasSuffix(name, ".jsonl.seal") && !strings.HasSuffix(name, ".jsonl.seal.tmp") {
			continue
		}
		day, _, _ := strings.Cut(name, ".jsonl")
		if _, err := time.Parse("2006-01-02", day); err == nil && day < cutoff {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			if len(s.integrityKey) > 0 {
				if err := os.Remove(path + ".seal"); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
		}
	}
	return nil
}
func (s *Store) Prune() error { s.mu.Lock(); defer s.mu.Unlock(); return s.pruneLocked() }

func (s *Store) Append(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.pruneLocked(); err != nil {
		s.lastError = true
		return err
	}
	now := s.now().UTC()
	e.Time = now.Format(time.RFC3339Nano)
	// Callers supply a registered route template, never a raw URL or query.
	if len(e.Route) > 200 {
		e.Route = "[unmatched]"
	}
	if s.maskedIP {
		e.Peer = MaskPeer(e.Peer)
	} else if !s.fullIP {
		e.Peer = ""
	}
	path := s.dayPath(now.Format("2006-01-02"))
	if len(s.integrityKey) > 0 {
		previous, err := s.verifyLocked(path)
		if err != nil {
			s.lastError = true
			return err
		}
		e.PreviousHash = previous
		e.Hash = s.digest(e)
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if s.strict && len(data) > 1024*1024 {
		s.lastError = true
		return errors.New("audit record too large")
	}
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			s.lastError = true
			return errors.New("invalid security log file")
		}
		if info.Size()+int64(len(data)) > s.maxBytes {
			s.dropped++
			if s.strict {
				s.lastError = true
				return errors.New("audit journal capacity reached")
			}
			return nil
		}
	} else if !os.IsNotExist(err) {
		s.lastError = true
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		s.lastError = true
		return err
	}
	_, err = file.Write(data)
	if err == nil && s.strict {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil && len(s.integrityKey) > 0 {
		err = s.writeSealLocked(path, e.Hash)
	}
	s.lastError = err != nil
	return err
}

// Read is bounded to one day's capped file and returns at most 100 records.
func (s *Store) Read(day string, offset int) (Page, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	page := Page{FullIP: s.fullIP, IntegrityChecked: len(s.integrityKey) > 0, Items: []Entry{}, RetentionDays: s.days, MaskedIP: s.maskedIP, WriteError: s.lastError, Dropped: s.dropped}
	if _, err := time.Parse("2006-01-02", day); err != nil || offset < 0 {
		return page, errors.New("invalid log query")
	}
	if err := s.pruneLocked(); err != nil {
		return page, err
	}
	path := s.dayPath(day)
	if _, err := s.verifyLocked(path); err != nil {
		return page, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return page, nil
	}
	if err != nil {
		return page, err
	}
	if !info.Mode().IsRegular() || info.Size() > s.maxBytes {
		return page, errors.New("invalid log file")
	}
	f, err := os.Open(path)
	if err != nil {
		return page, err
	}
	defer f.Close()
	entries := []Entry{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 2*1024*1024)
	for scanner.Scan() {
		var e Entry
		if json.Unmarshal(scanner.Bytes(), &e) == nil {
			entries = append(entries, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return page, err
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Time > entries[j].Time })
	if offset >= len(entries) {
		return page, nil
	}
	end := offset + 100
	if end > len(entries) {
		end = len(entries)
	}
	page.Items = entries[offset:end]
	if end < len(entries) {
		page.Next = end
	}
	return page, nil
}

func (s *Store) RetentionDays() int { return s.days }
func (s *Store) MaskedIP() bool     { return s.maskedIP }
func (s *Store) String() string     { return fmt.Sprintf("security log: %d days", s.days) }

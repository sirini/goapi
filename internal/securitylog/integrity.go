package securitylog

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Audit journals have independent retention and refuse further admin work if records cannot be written.
// The legacy days argument is ignored: administrator journals never expire automatically.
func NewAudit(dir string, _ int) (*Store, error) {
	s, err := New(dir, 0, 100*1024*1024, false)
	if err != nil {
		return nil, err
	}
	keyPath := filepath.Join(dir, ".integrity-key")
	key, err := os.ReadFile(keyPath)
	if os.IsNotExist(err) {
		files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		if len(files) > 0 {
			return nil, errors.New("audit key missing for existing records")
		}
		key = make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		f, openErr := os.OpenFile(keyPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if openErr != nil {
			return nil, openErr
		}
		_, err = f.Write(key)
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}
	if err != nil || len(key) != 32 {
		return nil, errors.New("invalid audit integrity key")
	}
	s.integrityKey = key
	s.fullIP = true
	s.strict = true
	return s, nil
}
func (s *Store) digest(e Entry) string {
	e.Hash = ""
	raw, _ := json.Marshal(e)
	var canonical map[string]json.RawMessage
	_ = json.Unmarshal(raw, &canonical)
	data, _ := json.Marshal(canonical)
	mac := hmac.New(sha256.New, s.integrityKey)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify the complete bounded journal, including order. Corrupt data is never silently exported.
func (s *Store) verifyLocked(path string) (string, error) {
	if len(s.integrityKey) == 0 {
		return "", nil
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		if _, sealErr := os.Stat(path + ".seal"); !os.IsNotExist(sealErr) {
			return "", errors.New("audit data missing")
		}
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() > s.maxBytes {
		return "", errors.New("invalid audit size")
	}
	previous := ""
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 2*1024*1024)
	for scanner.Scan() {
		var e Entry
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&e) != nil || e.PreviousHash != previous || !hmac.Equal([]byte(e.Hash), []byte(s.digest(e))) {
			return "", errors.New("audit integrity check failed")
		}
		previous = e.Hash
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	seal, err := os.ReadFile(path + ".seal")
	// Read-only compatibility for an interrupted deletion from the earlier implementation.
	// Never rename, delete, or rewrite the data or pending seal during verification.
	if os.IsNotExist(err) {
		pending := path + ".seal.deleting"
		if info, statErr := os.Lstat(pending); statErr == nil && info.Mode().IsRegular() && info.Size() <= 1024 {
			seal, err = os.ReadFile(pending)
		}
	}
	if err != nil {
		return "", errors.New("audit seal missing")
	}
	parts := strings.Split(string(seal), ".")
	if len(parts) != 2 || parts[0] != previous || !hmac.Equal([]byte(parts[1]), []byte(s.sealDigest(path, previous))) {
		return "", errors.New("audit seal mismatch")
	}
	return previous, nil
}

func (s *Store) sealDigest(path, hash string) string {
	mac := hmac.New(sha256.New, s.integrityKey)
	mac.Write([]byte(filepath.Base(path) + "|" + hash))
	return hex.EncodeToString(mac.Sum(nil))
}
func (s *Store) writeSealLocked(path, hash string) error {
	seal := []byte(hash + "." + s.sealDigest(path, hash))
	temp := path + ".seal.tmp"
	f, err := os.OpenFile(temp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(seal)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temp, path+".seal")
}

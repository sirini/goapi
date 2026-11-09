package securitylog

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

func csvCell(value string) string {
	if strings.ContainsAny(strings.TrimLeft(value, " "), "\r\n") {
		value = strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")
	}
	trimmed := strings.TrimLeft(value, " ")
	if trimmed != "" && strings.ContainsRune("=+-@\t\r", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}

func (s *Store) Export(day string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return nil, errors.New("invalid date")
	}
	if err := s.pruneLocked(); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	// BOM lets common spreadsheet software read UTF-8 correctly.
	buffer.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"time_utc", "method", "route_template", "http_status", "duration_ms", "actor_uid", "network_peer", "subject_uids", "phase", "integrity_hash", "export_purpose", "request_id", "result_code", "peer_source", "reference_date"})
	path := s.dayPath(day)
	if _, err := s.verifyLocked(path); err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() || info.Size() > s.maxBytes {
			return nil, errors.New("invalid log file")
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 4096), 2*1024*1024)
		for scanner.Scan() {
			var e Entry
			if json.Unmarshal(scanner.Bytes(), &e) != nil {
				continue
			}
			if err := writer.Write([]string{csvCell(e.Time), csvCell(e.Method), csvCell(e.Route), strconv.Itoa(e.Status), strconv.FormatInt(e.DurationMS, 10), strconv.FormatUint(uint64(e.Actor), 10), csvCell(e.Peer), csvCell(subjectCSV(e.Subjects)), csvCell(e.Phase), csvCell(e.Hash), csvCell(e.Purpose), csvCell(e.RequestID), resultCodeCSV(e.ResultCode), csvCell(e.PeerSource), csvCell(e.Reference)}); err != nil {
				return nil, err
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	writer.Flush()
	return buffer.Bytes(), writer.Error()
}

func subjectCSV(ids []uint) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatUint(uint64(id), 10)
	}
	return strings.Join(parts, ";")
}

func resultCodeCSV(code *int) string {
	if code == nil {
		return ""
	}
	return strconv.Itoa(*code)
}

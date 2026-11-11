package securitylog

import (
	"bytes"
	"encoding/csv"
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

func (s *Store) Export(day string) ([]byte, error) { return s.ExportInLocation(day, time.UTC) }

func (s *Store) ExportInLocation(day string, loc *time.Location) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.dateEntriesLocked(day, loc)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&buffer)
	timeHeader := "time_utc"
	if loc == KST {
		timeHeader = "time_kst"
	}
	_ = writer.Write([]string{timeHeader, "method", "route_template", "http_status", "duration_ms", "actor_uid", "network_peer", "subject_uids", "phase", "integrity_hash", "export_purpose", "request_id", "result_code", "peer_source", "reference_date"})
	for _, e := range entries {
		timestamp := e.Time
		if loc == KST {
			t, err := time.Parse(time.RFC3339Nano, e.Time)
			if err != nil {
				return nil, err
			}
			timestamp = t.In(KST).Format("2006-01-02 15:04:05")
		}
		if err := writer.Write([]string{csvCell(timestamp), csvCell(e.Method), csvCell(e.Route), strconv.Itoa(e.Status), strconv.FormatInt(e.DurationMS, 10), strconv.FormatUint(uint64(e.Actor), 10), csvCell(e.Peer), csvCell(subjectCSV(e.Subjects)), csvCell(e.Phase), csvCell(e.Hash), csvCell(e.Purpose), csvCell(e.RequestID), resultCodeCSV(e.ResultCode), csvCell(e.PeerSource), csvCell(e.Reference)}); err != nil {
			return nil, err
		}
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

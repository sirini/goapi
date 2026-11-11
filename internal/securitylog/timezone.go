package securitylog

import (
	"errors"
	"time"
)

// KST has a fixed UTC+09 offset; no host timezone or tzdata installation is required.
var KST = time.FixedZone("Asia/Seoul", 9*60*60)

func queryLocation(value string) (*time.Location, error) {
	switch value {
	case "", "UTC":
		return time.UTC, nil
	case "Asia/Seoul":
		return KST, nil
	default:
		return nil, errors.New("unsupported timezone")
	}
}

func utcDays(start, end time.Time) []string {
	first := start.UTC()
	first = time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)
	last := end.Add(-time.Nanosecond).UTC().Format("2006-01-02")
	days := []string{}
	for d := first; d.Format("2006-01-02") <= last; d = d.AddDate(0, 0, 1) {
		days = append(days, d.Format("2006-01-02"))
	}
	return days
}

func (s *Store) dateEntriesLocked(day string, loc *time.Location) ([]Entry, error) {
	start, err := time.ParseInLocation("2006-01-02", day, loc)
	if err != nil {
		return nil, err
	}
	if err := s.pruneLocked(); err != nil {
		return nil, err
	}
	if loc == time.UTC {
		return s.entriesLocked(day)
	}
	end := start.AddDate(0, 0, 1)
	entries := []Entry{}
	for _, utcDay := range utcDays(start, end) {
		rows, err := s.entriesLocked(utcDay)
		if err != nil {
			return nil, err
		}
		for _, e := range rows {
			t, err := time.Parse(time.RFC3339Nano, e.Time)
			if err != nil {
				return nil, err
			}
			if !t.Before(start) && t.Before(end) {
				entries = append(entries, e)
			}
		}
	}
	return entries, nil
}

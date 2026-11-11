package securitylog

import (
	"bufio"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type dayCount struct {
	size      int64
	modified  time.Time
	count     int
	kstCounts map[string]int
}

// NewVisits deliberately has no expiry setting, including for existing installations.
func NewVisits(dir string) (*Store, error) {
	s, err := New(dir, 0, 100*1024*1024, false)
	if err == nil {
		s.fullIP = true
	}
	return s, err
}

var Visits *Store
var pagePath = regexp.MustCompile(`^/[a-zA-Z0-9/%_-]*$`)

// Page paths exclude query strings, fragments and arbitrary text.
func VisitHandler(canAuthenticate func(uint) bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Cache-Control", "no-store")
		var payload struct {
			Path string `json:"path"`
		}
		if c.Bind().Body(&payload) != nil || len(payload.Path) > 200 || !pagePath.MatchString(payload.Path) || strings.HasPrefix(payload.Path, "//") {
			return c.SendStatus(400)
		}
		// Search terms and password reset codes can appear in path segments.
		decoded, decodeErr := url.PathUnescape(payload.Path)
		if decodeErr != nil || strings.ContainsAny(decoded, "?#\\\r\n\t") || strings.HasPrefix(decoded, "//") {
			return c.SendStatus(400)
		}
		payload.Path = decoded
		parts := strings.Split(decoded, "/")
		for index, part := range parts {
			if strings.EqualFold(part, "search") || strings.EqualFold(part, "change-password") && index > 0 && strings.EqualFold(parts[index-1], "auth") {
				payload.Path = strings.Join(parts[:index+1], "/")
				break
			}
		}

		actor := uint(0)
		if uid := utils.ExtractUserUid(c.Get(models.AUTH_KEY)); uid > 0 && canAuthenticate != nil && canAuthenticate(uint(uid)) {
			actor = uint(uid)
		} else if c.Get(models.AUTH_KEY) != "" {
			return c.SendStatus(401)
		}
		c.Locals("securityLogActor", actor)
		if Visits == nil {
			return c.SendStatus(503)
		}
		entry := Entry{Method: "PAGE", Route: payload.Path, Actor: actor, Peer: RequestPeer(c), PeerSource: PeerSource(c), Status: 200}
		if err := Visits.Append(entry); err != nil {
			return c.SendStatus(503)
		}
		adminPath := strings.ToLower(payload.Path)
		if actor == 1 && (adminPath == "/admin" || strings.HasPrefix(adminPath, "/admin/")) {
			if Audit == nil {
				return c.SendStatus(503)
			}
			entry.Phase = "page-visit"
			if err := Audit.Append(entry); err != nil {
				return c.SendStatus(503)
			}
		}
		return c.JSON(fiber.Map{"success": true})
	}
}

func VisitsReadHandler(c fiber.Ctx) error { return readStore(c, Visits) }

// Counts are page views, not unique people. Only retained days contribute.
func VisitStatisticsHandler(c fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	if Visits == nil {
		return c.SendStatus(503)
	}
	loc, err := queryLocation(c.Query("timezone"))
	if err != nil {
		return c.SendStatus(400)
	}
	now := Visits.now().In(loc)
	year := c.Query("year", now.Format("2006"))
	start, err := time.ParseInLocation("2006", year, loc)
	if err != nil {
		return c.SendStatus(400)
	}
	unit := c.Query("unit", "month")
	if unit != "day" && unit != "month" && unit != "year" {
		return c.SendStatus(400)
	}
	counts := map[string]int{}
	cutoff := now.AddDate(0, 0, -Visits.days+1).Format("2006-01-02")
	for d := start; d.Year() == start.Year(); d = d.AddDate(0, 0, 1) {
		day := d.Format("2006-01-02")
		if Visits.days > 0 && day < cutoff || day > now.Format("2006-01-02") {
			continue
		}
		key := day
		if unit == "month" {
			key = day[:7]
		}
		if unit == "year" {
			key = year
		}
		counts[key] += 0
		count, err := Visits.countDayInLocation(day, loc)
		if err != nil {
			return c.SendStatus(503)
		}
		counts[key] += count

	}
	return c.JSON(fiber.Map{"manualDeletionOnly": Visits.days == 0, "counts": counts, "retentionDays": Visits.days, "unit": unit, "year": year, "metric": "pageviews", "timezone": loc.String()})
}

func (s *Store) countDay(day string) (int, error) { return s.countDayInLocation(day, time.UTC) }

func (s *Store) countDayInLocation(day string, loc *time.Location) (int, error) {
	start, err := time.ParseInLocation("2006-01-02", day, loc)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.pruneLocked(); err != nil {
		return 0, err
	}
	count := 0
	for _, utcDay := range utcDays(start, start.AddDate(0, 0, 1)) {
		cached, err := s.countFileLocked(utcDay)
		if err != nil {
			return 0, err
		}
		if loc == KST {
			count += cached.kstCounts[day]
		} else {
			count += cached.count
		}
	}
	return count, nil
}

func (s *Store) countFileLocked(day string) (dayCount, error) {
	path := s.dayPath(day)
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return dayCount{}, nil
	}
	if err != nil {
		return dayCount{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > s.maxBytes {
		return dayCount{}, fiber.ErrServiceUnavailable
	}
	if cached, ok := s.countCache[day]; ok && cached.size == info.Size() && cached.modified.Equal(info.ModTime()) {
		return cached, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return dayCount{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 2*1024*1024)
	result := dayCount{size: info.Size(), modified: info.ModTime(), kstCounts: map[string]int{}}
	for scanner.Scan() {
		var entry Entry
		if json.Unmarshal(scanner.Bytes(), &entry) == nil {
			result.count++
			if stamp, err := time.Parse(time.RFC3339Nano, entry.Time); err == nil {
				result.kstCounts[stamp.In(KST).Format("2006-01-02")]++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return dayCount{}, err
	}
	if s.countCache == nil {
		s.countCache = map[string]dayCount{}
	}
	cutoff := s.now().UTC().AddDate(0, 0, -s.days+1).Format("2006-01-02")
	for key := range s.countCache {
		if s.days > 0 && key < cutoff || len(s.countCache) >= 800 {
			delete(s.countCache, key)
		}
	}
	s.countCache[day] = result
	return result, nil
}

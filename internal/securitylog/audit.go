package securitylog

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/netip"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
)

var Audit *Store

func RequestPeer(c fiber.Ctx) string {
	peer, err := netip.ParseAddr(c.IP())
	if err != nil {
		return ""
	}
	peer = peer.Unmap()
	for _, value := range strings.Split(configs.Env.TrustedProxyCIDRs, ",") {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil || prefix.Bits() == 0 || !prefix.Contains(peer) {
			continue
		}
		forwarded, err := netip.ParseAddr(strings.TrimSpace(c.Get("X-Real-IP")))
		if err == nil {
			return forwarded.Unmap().String()
		}
	}
	return peer.String()
}

// Only numeric subject identifiers are retained; values such as email, name or search terms are never copied.
func subjectIDs(c fiber.Ctx, response bool) []uint {
	ids := map[uint]bool{}
	add := func(value string) {
		n, err := strconv.ParseUint(value, 10, 32)
		if err == nil && n > 0 {
			ids[uint(n)] = true
		}
	}
	for _, key := range []string{"userUid", "targetUserUid", "user_uid"} {
		add(c.Query(key))
		add(c.FormValue(key))
	}
	var walk func(any, bool)
	walk = func(value any, userObject bool) {
		switch v := value.(type) {
		case map[string]any:
			for key, item := range v {
				if key == "userUid" || key == "targetUserUid" || key == "user_uid" || key == "uid" && userObject {
					if n, ok := item.(json.Number); ok {
						add(string(n))
					}
				} else if key == "result" || key == "items" || key == "item" || key == "users" {
					walk(item, userObject)
				}
			}
		case json.Number:
			if userObject {
				add(string(v))
			}
		case []any:
			for _, item := range v {
				walk(item, userObject)
			}
		}
	}
	var value any
	if !response && !strings.Contains(c.Get("Content-Type"), "application/json") {
		result := []uint{}
		for id := range ids {
			result = append(result, id)
		}
		sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
		return result
	}
	if response && !strings.Contains(string(c.Response().Header.ContentType()), "application/json") {
		return nil
	}
	data := c.Body()
	if response {
		data = c.Response().Body()
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if decoder.Decode(&value) == nil {
		walk(value, strings.Contains(c.Path(), "/admin/user/"))
	}
	result := []uint{}
	for id := range ids {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// Installed after active-root authorization. An intent record must be written before access is allowed.
func AuditMiddleware(c fiber.Ctx) error {
	if Audit == nil {
		return c.Status(503).JSON(fiber.Map{"error": "administrator audit unavailable"})
	}
	actor, _ := c.Locals("securityLogActor").(uint)
	method := c.Method()
	peer := RequestPeer(c)
	started := time.Now()
	var randomID [16]byte
	if _, err := rand.Read(randomID[:]); err != nil {
		return c.SendStatus(503)
	}
	requestID := hex.EncodeToString(randomID[:])
	reference := ""
	if value := c.Query("date"); value != "" {
		if _, err := time.Parse("2006-01-02", value); err == nil {
			reference = value
		}
	}
	purpose := ""
	if c.Query("format") == "csv" {
		purpose = c.Query("purpose")
		if purpose != "privacy-review" && purpose != "incident-response" && purpose != "authority-submission" {
			return c.Status(400).JSON(fiber.Map{"error": "export purpose required"})
		}
	}
	before := Entry{Reference: reference, RequestID: requestID, Purpose: purpose, Method: method, Route: "/admin/*", Actor: actor, Peer: peer, PeerSource: PeerSource(c), Subjects: subjectIDs(c, false), Phase: "intent"}
	if err := Audit.Append(before); err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "administrator audit could not be recorded"})
	}
	err := c.Next()
	status := c.Response().StatusCode()
	if err != nil {
		status = 500
		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
		}
	}
	route := c.Route().Path
	if c.Query("format") == "csv" {
		route += "#csv-export"
	}
	var responseCode struct {
		Code *int `json:"code"`
	}
	_ = json.Unmarshal(c.Response().Body(), &responseCode)
	after := Entry{Reference: reference, RequestID: requestID, ResultCode: responseCode.Code, Purpose: purpose, Method: method, Route: route, Actor: actor, Peer: peer, PeerSource: PeerSource(c), Subjects: subjectIDs(c, true), Phase: "completed", Status: status, DurationMS: time.Since(started).Milliseconds()}
	if appendErr := Audit.Append(after); appendErr != nil {
		c.Response().ResetBody()
		c.Set("Content-Type", "application/json")
		return c.Status(503).JSON(fiber.Map{"error": "operation may have completed; audit outcome could not be recorded"})
	}
	return err
}

func AuditReadHandler(c fiber.Ctx) error { return readStore(c, Audit) }

func AuditReviewHandler(c fiber.Ctx) error {
	month := c.Query("month")
	if month > time.Now().UTC().Format("2006-01") {
		return c.SendStatus(400)
	}
	if _, err := time.Parse("2006-01", month); err != nil {
		return c.SendStatus(400)
	}
	if Audit != nil {
		files, err := filepath.Glob(filepath.Join(Audit.dir, month+"-*.jsonl"))
		if err != nil {
			return c.SendStatus(503)
		}
		for _, path := range files {
			Audit.mu.Lock()
			_, err = Audit.verifyLocked(path)
			Audit.mu.Unlock()
			if err != nil {
				return c.Status(503).JSON(fiber.Map{"error": "audit integrity check failed"})
			}
		}
	}
	actor, _ := c.Locals("securityLogActor").(uint)
	if Audit == nil || Audit.Append(Entry{Method: "REVIEW", Route: "/admin/system/audit/review/" + month, Actor: actor, Peer: RequestPeer(c), PeerSource: PeerSource(c), Phase: "review-confirmed"}) != nil {
		return c.SendStatus(503)
	}
	c.Set("Cache-Control", "no-store")
	return c.JSON(fiber.Map{"success": true, "month": month})
}

func PeerSource(c fiber.Ctx) string {
	peer, err := netip.ParseAddr(c.IP())
	if err != nil {
		return "unknown"
	}
	if RequestPeer(c) != peer.Unmap().String() {
		return "trusted-proxy"
	}
	return "direct-peer"
}

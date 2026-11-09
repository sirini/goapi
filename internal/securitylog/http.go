package securitylog

import (
	"github.com/gofiber/fiber/v3"
	"strconv"
	"strings"
	"time"
)

// Default is initialized before serving requests; files are never publicly served.
var Default *Store

func Middleware(c fiber.Ctx) error {
	started := time.Now()
	err := c.Next()
	if Default == nil {
		return err
	}
	route := c.Route().Path
	// Only mutations, authentication and HTTP failures. Ordinary browsing is not tracked.
	status := c.Response().StatusCode()
	if err != nil {
		status = 500
		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
		}
	}
	if c.Method() == "GET" && status < 400 && !strings.Contains(route, "/admin/") {
		return err
	}
	if !strings.HasPrefix(route, "/") || route == "/" {
		route = "[unmatched]"
	}
	if strings.HasSuffix(route, "/system/logs") && c.Query("format") == "csv" {
		route += "#csv-export"
	}
	actor, _ := c.Locals("securityLogActor").(uint)
	_ = Default.Append(Entry{Method: c.Method(), Route: route, Status: status, DurationMS: time.Since(started).Milliseconds(), Actor: actor, Peer: RequestPeer(c), PeerSource: PeerSource(c)})
	return err
}

func ReadHandler(c fiber.Ctx) error { return readStore(c, Default) }

func readStore(c fiber.Ctx, store *Store) error {
	if store == nil {
		return c.Status(503).JSON(fiber.Map{"error": "security log unavailable"})
	}
	c.Set("Cache-Control", "no-store")
	if c.Query("format") == "csv" {
		day := c.Query("date", time.Now().UTC().Format("2006-01-02"))
		data, err := store.Export(day)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "log export unavailable"})
		}
		c.Set("Content-Type", "text/csv; charset=utf-8")
		prefix := "nubo-security-"
		if store == Audit {
			prefix = "nubo-audit-"
		}
		c.Set("Content-Disposition", `attachment; filename="`+prefix+day+`.csv"`)
		return c.Send(data)
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 {
		return c.SendStatus(400)
	}
	day := c.Query("date", time.Now().UTC().Format("2006-01-02"))
	page, err := store.Read(day, offset)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "log query unavailable"})
	}
	c.Set("Cache-Control", "no-store")
	return c.JSON(page)
}

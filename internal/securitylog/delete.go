package securitylog

import (
	"crypto/hmac"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// DeleteDay removes one completed UTC day, never the live journal or its key.
// The pending seal allows an explicitly retried deletion to recover after interruption.
func (s *Store) DeleteDay(day string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := time.Parse("2006-01-02", day); err != nil || day >= s.now().UTC().Format("2006-01-02") {
		return false, errors.New("only past UTC days can be deleted")
	}
	path := s.dayPath(day)
	pending := path + ".seal.deleting"
	if len(s.integrityKey) > 0 {
		if info, err := os.Lstat(pending); err == nil {
			if !info.Mode().IsRegular() || info.Size() > 1024 {
				return false, errors.New("invalid pending deletion")
			}
			data, err := os.ReadFile(pending)
			if err != nil {
				return false, err
			}
			parts := strings.Split(string(data), ".")
			if len(parts) != 2 || !hmac.Equal([]byte(parts[1]), []byte(s.sealDigest(path, parts[0]))) {
				return false, errors.New("invalid pending seal")
			}
			if _, err := os.Lstat(path + ".seal"); !os.IsNotExist(err) {
				return false, errors.New("conflicting audit seals")
			}
			if _, err := os.Lstat(path); os.IsNotExist(err) {
				return true, os.Remove(pending)
			} else if err != nil {
				return false, err
			}
			if err := os.Rename(pending, path+".seal"); err != nil {
				return false, err
			}
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if _, err := s.verifyLocked(path); err != nil {
			return false, err
		}
		delete(s.countCache, day)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, errors.New("invalid log file")
	}
	if _, err := s.verifyLocked(path); err != nil {
		return false, err
	}
	if len(s.integrityKey) > 0 {
		info, err := os.Lstat(path + ".seal")
		if err != nil {
			return false, err
		}
		if !info.Mode().IsRegular() {
			return false, errors.New("invalid seal file")
		}
		if err := os.Rename(path+".seal", pending); err != nil {
			return false, err
		}
	}
	if err := os.Remove(path); err != nil {
		if len(s.integrityKey) > 0 {
			_ = os.Rename(pending, path+".seal")
		}
		return false, err
	}
	delete(s.countCache, day)
	if len(s.integrityKey) > 0 {
		if err := os.Remove(pending); err != nil {
			return true, err
		}
	}
	directory, err := os.Open(s.dir)
	if err != nil {
		return true, err
	}
	defer directory.Close()
	return true, directory.Sync()
}

// Mounted behind active-root authorization and the fail-closed audit middleware.
func DeleteRecordsHandler(c fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	kind, day := c.Params("kind"), c.Query("date")
	var body struct {
		ConfirmDate string `json:"confirmDate"`
		Reason      string `json:"reason"`
	}
	if c.Bind().Body(&body) != nil || body.ConfirmDate != day {
		return c.SendStatus(400)
	}
	if body.Reason != "retention-policy" && body.Reason != "privacy-request" && body.Reason != "test-cleanup" {
		return c.SendStatus(400)
	}
	if _, err := time.Parse("2006-01-02", day); err != nil || day >= time.Now().UTC().Format("2006-01-02") {
		return c.SendStatus(400)
	}
	store := Visits
	if kind == "audit" {
		store = Audit
	} else if kind != "visits" {
		return c.SendStatus(400)
	}
	if store == nil || Audit == nil {
		return c.SendStatus(503)
	}
	actor, _ := c.Locals("securityLogActor").(uint)
	if actor != 1 {
		return c.SendStatus(403)
	}
	requestID, _ := c.Locals("securityLogRequestID").(string)
	entry := Entry{Method: "DELETE", Route: "/admin/system/records/" + kind, Reference: day, Purpose: body.Reason, Actor: actor, Peer: RequestPeer(c), PeerSource: PeerSource(c), RequestID: requestID, Phase: "delete-intent"}
	if err := Audit.Append(entry); err != nil {
		return c.SendStatus(503)
	}
	deleted, err := store.DeleteDay(day)
	entry.Phase = "delete-completed"
	entry.Status = 200
	if err != nil {
		entry.Phase = "delete-failed"
		entry.Status = 503
	}
	if appendErr := Audit.Append(entry); appendErr != nil {
		return c.Status(503).JSON(fiber.Map{"error": "deletion may have completed; audit outcome unavailable"})
	}
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "deletion incomplete; check records before retrying"})
	}
	return c.JSON(fiber.Map{"success": true, "deleted": deleted, "date": day, "kind": kind})
}

package handlers

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
)

//go:embed api-contract-version.txt
var embeddedAPIContractVersion string

var apiContractVersion = strings.TrimSpace(embeddedAPIContractVersion)

type StatusHandler interface {
	HealthHandler(c fiber.Ctx) error
	ReadyHandler(c fiber.Ctx) error
	VersionHandler(c fiber.Ctx) error
}

type databasePinger interface {
	PingContext(context.Context) error
}

type NuboStatusHandler struct {
	db      databasePinger
	version string
}

func NewNuboStatusHandler(db *sql.DB) *NuboStatusHandler {
	return newNuboStatusHandler(db, configs.Env.Version)
}

func newNuboStatusHandler(db databasePinger, version string) *NuboStatusHandler {
	return &NuboStatusHandler{db: db, version: version}
}

func (h *NuboStatusHandler) HealthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "goapi"})
}

func (h *NuboStatusHandler) ReadyHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if h.db == nil || h.db.PingContext(ctx) != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "unavailable",
			"service": "goapi",
		})
	}
	return c.JSON(fiber.Map{"status": "ok", "service": "goapi"})
}

func (h *NuboStatusHandler) VersionHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":      "ok",
		"service":     "goapi",
		"version":     h.version,
		"apiContract": apiContractVersion,
	})
}

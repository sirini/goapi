package handlers

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/securitylog"
)

//go:embed api-contract-version.txt
var embeddedAPIContractVersion string

var apiContractVersion = strings.TrimSpace(embeddedAPIContractVersion)

type StatusHandler interface {
	PrivacyStatusHandler(c fiber.Ctx) error
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

// Only non-secret operational facts. Keys, identifiers, paths and hostnames are never exposed.
func (h *NuboStatusHandler) PrivacyStatusHandler(c fiber.Ctx) error {
	env := configs.Env
	ai := configs.GetImageDescriptionConfig()
	ai.Enabled = ai.Enabled && ai.MaxPerPost > 0
	providers := []string{}
	if strings.TrimSpace(env.ResendKey) != "" && strings.TrimSpace(env.ResendFromEmail) != "" {
		providers = append(providers, "Resend")
	}
	if len(configs.GetAppleClientIDs()) > 0 {
		providers = append(providers, "Apple")
	}
	if strings.TrimSpace(env.OAuthGoogleID) != "" || strings.TrimSpace(env.OAuthGoogleAndroidID) != "" {
		providers = append(providers, "Google")
	}
	if strings.TrimSpace(env.OAuthNaverID) != "" {
		providers = append(providers, "Naver")
	}
	if strings.TrimSpace(env.OAuthKakaoID) != "" {
		providers = append(providers, "Kakao")
	}
	if strings.TrimSpace(env.FirebaseProjectID) != "" && strings.TrimSpace(env.FirebaseCredentialsFile) != "" {
		providers = append(providers, "Firebase")
	}
	if ai.Enabled {
		providers = append(providers, "OpenAI")
	}
	auditDays := 0
	if securitylog.Audit != nil {
		auditDays = securitylog.Audit.RetentionDays()
	}
	days := 0
	masked := false
	if securitylog.Default != nil {
		days = securitylog.Default.RetentionDays()
		masked = securitylog.Default.MaskedIP()
	}
	visitDays := 0
	if securitylog.Visits != nil {
		visitDays = securitylog.Visits.RetentionDays()
	}
	c.Set("Cache-Control", "no-store")
	return c.JSON(fiber.Map{"visitLogEnabled": securitylog.Visits != nil, "auditLogEnabled": securitylog.Audit != nil, "visitLogManualDeletionOnly": securitylog.Visits != nil && visitDays == 0, "auditLogManualDeletionOnly": securitylog.Audit != nil && auditDays == 0, "visitLogRetentionDays": visitDays, "auditLogRetentionDays": auditDays, "providers": providers, "imageDescriptionEnabled": ai.Enabled, "securityLogEnabled": securitylog.Default != nil, "securityLogRetentionDays": days, "securityLogMaskedIP": masked})
}

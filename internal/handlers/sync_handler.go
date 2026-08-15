package handlers

import (
	"crypto/subtle"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

func syncKeyMatches(provided, configured string) bool {
	if provided == "" || configured == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(configured)) == 1
}

func configuredSyncKey() string {
	if configs.Env.SyncSecretKey != "" {
		return configs.Env.SyncSecretKey
	}
	return configs.Env.JWTSecretKey
}

type SyncHandler interface {
	SyncPostHandler(c fiber.Ctx) error
}

type NuboSyncHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewNuboSyncHandler(service *services.Service) *NuboSyncHandler {
	return &NuboSyncHandler{service: service}
}

// (허용된) 다른 곳으로 이 곳의 게시글들을 동기화 할 수 있도록 데이터 출력
func (h *NuboSyncHandler) SyncPostHandler(c fiber.Ctx) error {
	key := c.Get("X-Sync-Key")
	if key == "" {
		key = c.FormValue("key")
	}
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil || bunch < 1 || bunch > 100 {
		return utils.Err(c, "Invalid limit, not a valid number", models.CODE_INVALID_PARAMETER)
	}

	if !syncKeyMatches(key, configuredSyncKey()) {
		return utils.Err(c, "Invalid key, unauthorized access", models.CODE_INVALID_PARAMETER)
	}

	result := h.service.Sync.GetLatestPosts(uint(bunch))
	return utils.Ok(c, result)
}

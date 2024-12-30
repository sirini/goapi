package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/utils"
)

type SyncHandler interface {
	SyncPostHandler(c fiber.Ctx) error
}

type TsboardSyncHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardSyncHandler(service *services.Service) *TsboardSyncHandler {
	return &TsboardSyncHandler{service: service}
}

// (허용된) 다른 곳으로 이 곳의 게시글들을 동기화 할 수 있도록 데이터 출력
func (h *TsboardSyncHandler) SyncPostHandler(c fiber.Ctx) error {
	key := c.FormValue("key")
	bunch, err := strconv.ParseUint(c.FormValue("limit"), 10, 32)
	if err != nil || bunch < 1 || bunch > 100 {
		return utils.Err(c, "Invalid limit, not a valid number")
	}

	if key != configs.Env.JWTSecretKey {
		return utils.Err(c, "Invalid key, unauthorized access")
	}

	result := h.service.Sync.GetLatestPosts(uint(bunch))
	return utils.Ok(c, result)
}

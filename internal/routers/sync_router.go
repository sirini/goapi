package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
)

// 다른 서버에 이곳 데이터를 동기화 시킬 때 필요한 라우터 등록
func RegisterSyncRouters(api fiber.Router, h *handlers.Handler) {
	api.Get("/sync", h.Sync.SyncPostHandler)
}

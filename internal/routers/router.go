package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
)

// 라우터들 등록하기
func RegisterRouters(api fiber.Router, h *handlers.Handler) {
	RegisterAdminRouters(api, h)
	RegisterAuthRouters(api, h)
	RegisterBoardRouters(api, h)
	RegisterBlogRouters(api, h)
	RegisterChatRouters(api, h)
	RegisterCommentRouters(api, h)
	RegisterEditorRouters(api, h)
	RegisterHomeRouters(api, h)
	RegisterNotiRouters(api, h)
	RegisterSyncRouters(api, h)
	RegisterTradeRouters(api, h)
	RegisterUserRouters(api, h)
}

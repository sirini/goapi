package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

func RegisterPushRouters(api fiber.Router, h *handlers.Handler) {
	push := api.Group("/push", middlewares.JWTMiddleware(h.CanAuthenticate))
	push.Post("/device", h.Push.RegisterDeviceHandler)
	push.Delete("/device", h.Push.UnregisterDeviceHandler)
}

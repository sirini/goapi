package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 글작성 에디터와 상호작용할 때 필요한 라우터들 등록
func RegisterEditorRouters(api fiber.Router, h *handlers.Handler) {
	editor := api.Group("/editor")
	editor.Get("/config", h.Editor.GetEditorConfigHandler)

	editor.Get("/load/images", h.Editor.LoadInsertImageHandler, middlewares.JWTMiddleware())
	editor.Get("/load/post", h.Editor.LoadPostHandler, middlewares.JWTMiddleware())
	editor.Patch("/modify", h.Editor.ModifyPostHandler, middlewares.JWTMiddleware())
	editor.Delete("/remove/attached", h.Editor.RemoveAttachedFileHandler, middlewares.JWTMiddleware())
	editor.Delete("/remove/image", h.Editor.RemoveInsertImageHandler, middlewares.JWTMiddleware())
	editor.Get("/suggestion/title", h.Editor.SuggestionTitleHandler, middlewares.JWTMiddleware())
	editor.Get("/suggestion/tag", h.Editor.SuggestionHashtagHandler, middlewares.JWTMiddleware())
	editor.Post("/upload/images", h.Editor.UploadInsertImageHandler, middlewares.JWTMiddleware())
	editor.Post("/write", h.Editor.WritePostHandler, middlewares.JWTMiddleware())
}

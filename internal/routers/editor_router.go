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
	editor.Get("/load/thumbnail", h.Editor.LoadThumbnailImageHandler)

	protected := editor.Group("/", middlewares.JWTMiddleware())
	protected.Get("/load/images", h.Editor.LoadInsertImageHandler)
	protected.Get("/load/post", h.Editor.LoadPostHandler)
	protected.Patch("/modify", h.Editor.ModifyPostHandler)
	protected.Delete("/remove/attached", h.Editor.RemoveAttachedFileHandler)
	protected.Delete("/remove/image", h.Editor.RemoveInsertImageHandler)
	protected.Get("/suggestion/title", h.Editor.SuggestionTitleHandler)
	protected.Get("/suggestion/tag", h.Editor.SuggestionHashtagHandler)
	protected.Post("/upload/images", h.Editor.UploadInsertImageHandler)
	protected.Post("/write", h.Editor.WritePostHandler)
}

package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 관리화면과 상호작용에 필요한 라우터들 등록
func RegisterAdminRouters(api fiber.Router, h *handlers.Handler) {
	admin := api.Group("/admin")
	board := admin.Group("/board")
	general := board.Group("/general")

	general.Post("/add/category", h.Admin.AddBoardCategoryHandler, middlewares.AdminMiddleware())
	general.Get("/load", h.Admin.BoardGeneralLoadHandler, middlewares.AdminMiddleware())
	general.Patch("/change/group", h.Admin.ChangeBoardGroupHandler, middlewares.AdminMiddleware())
	general.Patch("/change/name", h.Admin.ChangeBoardNameHandler, middlewares.AdminMiddleware())
	general.Patch("/change/info", h.Admin.ChangeBoardInfoHandler, middlewares.AdminMiddleware())
	general.Patch("/change/type", h.Admin.ChangeBoardTypeHandler, middlewares.AdminMiddleware())
	general.Patch("/change/rows", h.Admin.ChangeBoardRowHandler, middlewares.AdminMiddleware())
	general.Patch("/change/width", h.Admin.ChangeBoardWidthHandler, middlewares.AdminMiddleware())
	general.Delete("/remove/category", h.Admin.RemoveBoardCategoryHandler, middlewares.AdminMiddleware())
	general.Patch("/use/category", h.Admin.UseBoardCategoryHandler, middlewares.AdminMiddleware())
}

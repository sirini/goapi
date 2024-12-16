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
	dashboard := admin.Group("/dashboard")

	bGeneral := board.Group("/general")
	bGeneral.Post("/add/category", h.Admin.AddBoardCategoryHandler, middlewares.AdminMiddleware())
	bGeneral.Get("/load", h.Admin.BoardGeneralLoadHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/group", h.Admin.ChangeBoardGroupHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/name", h.Admin.ChangeBoardNameHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/info", h.Admin.ChangeBoardInfoHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/type", h.Admin.ChangeBoardTypeHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/rows", h.Admin.ChangeBoardRowHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/width", h.Admin.ChangeBoardWidthHandler, middlewares.AdminMiddleware())
	bGeneral.Delete("/remove/category", h.Admin.RemoveBoardCategoryHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/use/category", h.Admin.UseBoardCategoryHandler, middlewares.AdminMiddleware())

	bPermission := board.Group("/permission")
	bPermission.Get("/load", h.Admin.BoardLevelLoadHandler, middlewares.AdminMiddleware())
	bPermission.Patch("/change/admin", h.Admin.ChangeBoardAdminHandler, middlewares.AdminMiddleware())
	bPermission.Patch("/update/levels", h.Admin.ChangeBoardLevelHandler, middlewares.AdminMiddleware())
	bPermission.Get("/candidates", h.Admin.GetAdminCandidatesHandler, middlewares.AdminMiddleware())

	bPoint := board.Group("/point")
	bPoint.Get("/load", h.Admin.BoardPointLoadHandler, middlewares.AdminMiddleware())
	bPoint.Patch("/update/points", h.Admin.ChangeBoardPointHandler, middlewares.AdminMiddleware())

	dGeneral := dashboard.Group("/general")
	dLoad := dGeneral.Group("/load")
	dLoad.Get("/item", h.Admin.DashboardItemLoadHandler, middlewares.AdminMiddleware())
	dLoad.Get("/latest", h.Admin.DashboardLatestLoadHandler, middlewares.AdminMiddleware())
	dLoad.Get("/statistic", h.Admin.DashboardStatisticLoadHandler, middlewares.AdminMiddleware())

}

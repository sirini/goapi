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
	group := admin.Group("/group")
	latest := admin.Group("/latest")
	report := admin.Group("/report")
	user := admin.Group("/user")

	bGeneral := board.Group("/general")

	bGeneral.Get("/load", h.Admin.BoardGeneralLoadHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/group", h.Admin.ChangeBoardGroupHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/name", h.Admin.ChangeBoardNameHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/info", h.Admin.ChangeBoardInfoHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/type", h.Admin.ChangeBoardTypeHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/rows", h.Admin.ChangeBoardRowHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/change/width", h.Admin.ChangeBoardWidthHandler, middlewares.AdminMiddleware())
	bGeneral.Post("/category/add", h.Admin.AddBoardCategoryHandler, middlewares.AdminMiddleware())
	bGeneral.Delete("/category/remove", h.Admin.RemoveBoardCategoryHandler, middlewares.AdminMiddleware())
	bGeneral.Patch("/category/use", h.Admin.UseBoardCategoryHandler, middlewares.AdminMiddleware())

	bPermission := board.Group("/permission")
	bPermission.Get("/load", h.Admin.BoardLevelLoadHandler, middlewares.AdminMiddleware())
	bPermission.Get("/candidates", h.Admin.GetAdminCandidatesHandler, middlewares.AdminMiddleware())
	bPermission.Patch("/change/admin", h.Admin.ChangeBoardAdminHandler, middlewares.AdminMiddleware())
	bPermission.Patch("/change/levels", h.Admin.ChangeBoardLevelHandler, middlewares.AdminMiddleware())

	bPoint := board.Group("/point")
	bPoint.Get("/load", h.Admin.BoardPointLoadHandler, middlewares.AdminMiddleware())
	bPoint.Patch("/update", h.Admin.ChangeBoardPointHandler, middlewares.AdminMiddleware())

	dashboard.Get("/item", h.Admin.DashboardItemLoadHandler, middlewares.AdminMiddleware())
	dashboard.Get("/latest", h.Admin.DashboardLatestLoadHandler, middlewares.AdminMiddleware())
	dashboard.Get("/statistic", h.Admin.DashboardStatisticLoadHandler, middlewares.AdminMiddleware())

	gGeneral := group.Group("/general")
	gGeneral.Get("/load", h.Admin.GroupGeneralLoadHandler, middlewares.AdminMiddleware())
	gGeneral.Get("/candidates", h.Admin.GetAdminCandidatesHandler, middlewares.AdminMiddleware())
	gGeneral.Get("/boardids", h.Admin.ShowSimilarBoardIdHandler, middlewares.AdminMiddleware())
	gGeneral.Patch("/changeadmin", h.Admin.ChangeGroupAdminHandler, middlewares.AdminMiddleware())
	gGeneral.Delete("/board/remove", h.Admin.RemoveBoardHandler, middlewares.AdminMiddleware())
	gGeneral.Post("/board/create", h.Admin.CreateBoardHandler, middlewares.AdminMiddleware())

	gList := group.Group("/list")
	gList.Get("/load", h.Admin.GroupListLoadHandler, middlewares.AdminMiddleware())
	gList.Get("/groupids", h.Admin.ShowSimilarGroupIdHandler, middlewares.AdminMiddleware())
	gList.Post("/create", h.Admin.CreateGroupHandler, middlewares.AdminMiddleware())
	gList.Delete("/remove", h.Admin.RemoveGroupHandler, middlewares.AdminMiddleware())
	gList.Put("/update", h.Admin.ChangeGroupIdHandler, middlewares.AdminMiddleware())

	latest.Get("/comment", h.Admin.LatestCommentLoadHandler, middlewares.AdminMiddleware())
	latest.Delete("/comment", h.Admin.RemoveCommentHandler, middlewares.AdminMiddleware())
	latest.Get("/post", h.Admin.LatestPostLoadHandler, middlewares.AdminMiddleware())
	latest.Delete("/post", h.Admin.RemovePostHandler, middlewares.AdminMiddleware())
	latest.Get("/search/comment", h.Admin.LatestCommentSearchHandler, middlewares.AdminMiddleware())
	latest.Get("/search/post", h.Admin.LatestPostSearchHandler, middlewares.AdminMiddleware())

	report.Get("/list", h.Admin.ReportListLoadHandler, middlewares.AdminMiddleware())
	report.Get("/search", h.Admin.ReportListSearchHandler, middlewares.AdminMiddleware())

	user.Get("/list", h.Admin.UserListLoadHandler, middlewares.AdminMiddleware())
	user.Get("/load", h.Admin.UserInfoLoadHandler, middlewares.AdminMiddleware())
	user.Patch("/modify", h.Admin.UserInfoModifyHandler, middlewares.AdminMiddleware())
}

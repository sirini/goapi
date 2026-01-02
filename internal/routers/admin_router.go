package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 관리화면과 상호작용에 필요한 라우터들 등록
func RegisterAdminRouters(api fiber.Router, h *handlers.Handler) {
	admin := api.Group("/admin", middlewares.AdminMiddleware())
	board := admin.Group("/board", middlewares.AdminMiddleware())
	dashboard := admin.Group("/dashboard", middlewares.AdminMiddleware())
	group := admin.Group("/group", middlewares.AdminMiddleware())
	latest := admin.Group("/latest", middlewares.AdminMiddleware())
	report := admin.Group("/report", middlewares.AdminMiddleware())
	user := admin.Group("/user", middlewares.AdminMiddleware())

	bGeneral := board.Group("/general")

	bGeneral.Get("/load", h.Admin.BoardGeneralLoadHandler)
	bGeneral.Patch("/change/group", h.Admin.ChangeBoardGroupHandler)
	bGeneral.Patch("/change/name", h.Admin.ChangeBoardNameHandler)
	bGeneral.Patch("/change/info", h.Admin.ChangeBoardInfoHandler)
	bGeneral.Patch("/change/type", h.Admin.ChangeBoardTypeHandler)
	bGeneral.Patch("/change/rows", h.Admin.ChangeBoardRowHandler)
	bGeneral.Patch("/change/width", h.Admin.ChangeBoardWidthHandler)
	bGeneral.Post("/category/add", h.Admin.AddBoardCategoryHandler)
	bGeneral.Delete("/category/remove", h.Admin.RemoveBoardCategoryHandler)
	bGeneral.Patch("/category/use", h.Admin.UseBoardCategoryHandler)

	bPermission := board.Group("/permission")
	bPermission.Get("/load", h.Admin.BoardLevelLoadHandler)
	bPermission.Get("/candidates", h.Admin.GetAdminCandidatesHandler)
	bPermission.Patch("/change/admin", h.Admin.ChangeBoardAdminHandler)
	bPermission.Patch("/change/levels", h.Admin.ChangeBoardLevelHandler)

	bPoint := board.Group("/point")
	bPoint.Get("/load", h.Admin.BoardPointLoadHandler)
	bPoint.Patch("/update", h.Admin.ChangeBoardPointHandler)

	dashboard.Get("/item", h.Admin.DashboardItemLoadHandler)
	dashboard.Get("/latest", h.Admin.DashboardLatestLoadHandler)
	dashboard.Get("/statistic", h.Admin.DashboardStatisticLoadHandler)

	gGeneral := group.Group("/general")
	gGeneral.Get("/load", h.Admin.GroupGeneralLoadHandler)
	gGeneral.Get("/candidates", h.Admin.GetAdminCandidatesHandler)
	gGeneral.Get("/boardids", h.Admin.ShowSimilarBoardIdHandler)
	gGeneral.Patch("/changeadmin", h.Admin.ChangeGroupAdminHandler)
	gGeneral.Delete("/board/remove", h.Admin.RemoveBoardHandler)
	gGeneral.Post("/board/create", h.Admin.CreateBoardHandler)

	gList := group.Group("/list")
	gList.Get("/load", h.Admin.GroupListLoadHandler)
	gList.Get("/groupids", h.Admin.ShowSimilarGroupIdHandler)
	gList.Post("/create", h.Admin.CreateGroupHandler)
	gList.Delete("/remove", h.Admin.RemoveGroupHandler)
	gList.Post("/update", h.Admin.ChangeGroupIdHandler)

	latest.Get("/comment", h.Admin.LatestCommentLoadHandler)
	latest.Delete("/comment", h.Admin.RemoveCommentHandler)
	latest.Get("/post", h.Admin.LatestPostLoadHandler)
	latest.Delete("/post", h.Admin.RemovePostHandler)
	latest.Get("/search/comment", h.Admin.LatestCommentSearchHandler)
	latest.Get("/search/post", h.Admin.LatestPostSearchHandler)

	report.Get("/list", h.Admin.ReportListLoadHandler)
	report.Get("/search", h.Admin.ReportListSearchHandler)

	user.Get("/list", h.Admin.UserListLoadHandler)
	user.Get("/load", h.Admin.UserInfoLoadHandler)
	user.Patch("/modify", h.Admin.UserInfoModifyHandler)
}

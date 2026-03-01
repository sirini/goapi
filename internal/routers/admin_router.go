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

	board.Get("/load", h.Admin.BoardGeneralLoadHandler)
	board.Post("/create", h.Admin.CreateBoardHandler)
	board.Post("/modify", h.Admin.ModifyBoardHandler)
	board.Delete("/remove", h.Admin.RemoveBoardHandler)
	board.Get("/candidates", h.Admin.GetAdminCandidatesHandler)

	dashboard.Get("/usage", h.Admin.DashboardUploadUsageHandler)
	dashboard.Get("/item", h.Admin.DashboardItemLoadHandler)
	dashboard.Get("/statistic", h.Admin.DashboardStatisticLoadHandler)

	group.Get("/load", h.Admin.GroupGeneralLoadHandler)
	group.Get("/candidates", h.Admin.GetAdminCandidatesHandler)
	group.Get("/boardids", h.Admin.ShowSimilarBoardIdHandler)
	group.Get("/list", h.Admin.GroupListLoadHandler)
	group.Get("/groupids", h.Admin.ShowSimilarGroupIdHandler)
	group.Post("/create", h.Admin.CreateGroupHandler)
	group.Delete("/remove", h.Admin.RemoveGroupHandler)
	group.Post("/update", h.Admin.ChangeGroupIdHandler)

	latest.Delete("/comment", h.Admin.RemoveCommentHandler)
	latest.Delete("/post", h.Admin.RemovePostHandler)
	latest.Get("/comments", h.Admin.LatestCommentSearchHandler)
	latest.Get("/posts", h.Admin.LatestPostSearchHandler)

	report.Get("/reports", h.Admin.ReportListSearchHandler)

	user.Post("/create", h.Admin.CreateUserHandler)
	user.Get("/list", h.Admin.UserListLoadHandler)
	user.Get("/load", h.Admin.UserInfoLoadHandler)
	user.Patch("/modify", h.Admin.UserInfoModifyHandler)
	user.Delete("/remove", h.Admin.RemoveUserHandler)
}

package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
)

// 관리화면과 상호작용에 필요한 라우터들 등록
func RegisterAdminRouters(api fiber.Router, h *handlers.Handler) {
	api.Get("/skin/settings", h.Admin.SkinSettingsLoadHandler)
	admin := api.Group("/admin", middlewares.AdminMiddleware(h.CanAuthenticate))
	board := admin.Group("/board")
	dashboard := admin.Group("/dashboard")
	group := admin.Group("/group")
	latest := admin.Group("/latest")
	mail := admin.Group("/mail")
	report := admin.Group("/report")
	user := admin.Group("/user")
	skin := admin.Group("/skin")
	system := admin.Group("/system")
	skin.Put("/setting", h.Admin.SkinSettingModifyHandler)
	system.Get("/mail", h.Admin.MailStatusHandler)
	mail.Get("/campaigns", h.Admin.MailCampaignListHandler)
	mail.Get("/campaign/:uid", h.Admin.MailCampaignLoadHandler)
	mail.Post("/preview", h.Admin.MailCampaignPreviewHandler)
	mail.Post("/campaign", h.Admin.MailCampaignSaveHandler)
	mail.Post("/campaign/:uid/test", h.Admin.MailCampaignTestHandler)
	mail.Post("/campaign/:uid/prepare", h.Admin.MailCampaignPrepareHandler)
	mail.Post("/campaign/:uid/send", h.Admin.MailCampaignSendHandler)

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
	group.Post("/admin", h.Admin.ChangeGroupAdminHandler)

	latest.Delete("/comment", h.Admin.RemoveCommentHandler)
	latest.Delete("/post", h.Admin.RemovePostHandler)
	latest.Get("/comments", h.Admin.LatestCommentSearchHandler)
	latest.Get("/posts", h.Admin.LatestPostSearchHandler)

	report.Get("/reports", h.Admin.ReportListSearchHandler)
	report.Put("/resolve", h.Admin.ReportResolveHandler)

	user.Post("/create", h.Admin.CreateUserHandler)
	user.Get("/list", h.Admin.UserListLoadHandler)
	user.Get("/load", h.Admin.UserInfoLoadHandler)
	user.Post("/modify", h.Admin.UserInfoModifyHandler)
	user.Delete("/remove", h.Admin.RemoveUserHandler)
}

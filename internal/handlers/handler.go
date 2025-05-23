package handlers

import "github.com/sirini/goapi/internal/services"

// 모든 핸들러들을 관리
type Handler struct {
	Admin   AdminHandler
	Auth    AuthHandler
	Board   BoardHandler
	Blog    BlogHandler
	Chat    ChatHandler
	Comment CommentHandler
	Editor  EditorHandler
	Home    HomeHandler
	Noti    NotiHandler
	OAuth2  OAuth2Handler
	Sync    SyncHandler
	Trade   TradeHandler
	User    UserHandler
}

// 모든 핸들러들을 생성
func NewHandler(s *services.Service) *Handler {
	return &Handler{
		Admin:   NewTsboardAdminHandler(s),
		Auth:    NewTsboardAuthHandler(s),
		Board:   NewTsboardBoardHandler(s),
		Blog:    NewTsboardBlogHandler(s),
		Chat:    NewTsboardChatHandler(s),
		Comment: NewTsboardCommentHandler(s),
		Editor:  NewTsboardEditorHandler(s),
		Home:    NewTsboardHomeHandler(s),
		Noti:    NewTsboardNotiHandler(s),
		OAuth2:  NewTsboardOAuth2Handler(s),
		Sync:    NewTsboardSyncHandler(s),
		Trade:   NewTsboardTradeHandler(s),
		User:    NewTsboardUserHandler(s),
	}
}

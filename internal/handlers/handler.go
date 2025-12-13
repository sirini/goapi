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
		Admin:   NewNuboAdminHandler(s),
		Auth:    NewNuboAuthHandler(s),
		Board:   NewNuboBoardHandler(s),
		Blog:    NewNuboBlogHandler(s),
		Chat:    NewNuboChatHandler(s),
		Comment: NewNuboCommentHandler(s),
		Editor:  NewNuboEditorHandler(s),
		Home:    NewNuboHomeHandler(s),
		Noti:    NewNuboNotiHandler(s),
		OAuth2:  NewNuboOAuth2Handler(s),
		Sync:    NewNuboSyncHandler(s),
		Trade:   NewNuboTradeHandler(s),
		User:    NewNuboUserHandler(s),
	}
}

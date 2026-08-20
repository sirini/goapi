package handlers

import (
	"database/sql"

	"github.com/sirini/goapi/internal/services"
)

// 모든 핸들러들을 관리
type Handler struct {
	CanAuthenticate func(uint) bool
	Admin           AdminHandler
	Auth            AuthHandler
	Board           BoardHandler
	Blog            BlogHandler
	Chat            ChatHandler
	Comment         CommentHandler
	Editor          EditorHandler
	Home            HomeHandler
	Noti            NotiHandler
	OAuth2          OAuth2Handler
	Push            PushHandler
	Status          StatusHandler
	Sync            SyncHandler
	Trade           TradeHandler
	User            UserHandler
}

// 모든 핸들러들을 생성
func NewHandler(s *services.Service, db *sql.DB) *Handler {
	return &Handler{
		CanAuthenticate: s.Auth.CanAuthenticate,
		Admin:           NewNuboAdminHandler(s),
		Auth:            NewNuboAuthHandler(s),
		Board:           NewNuboBoardHandler(s),
		Blog:            NewNuboBlogHandler(s),
		Chat:            NewNuboChatHandler(s),
		Comment:         NewNuboCommentHandler(s),
		Editor:          NewNuboEditorHandler(s),
		Home:            NewNuboHomeHandler(s),
		Noti:            NewNuboNotiHandler(s),
		OAuth2:          NewNuboOAuth2Handler(s),
		Push:            NewNuboPushHandler(s),
		Status:          NewNuboStatusHandler(db),
		Sync:            NewNuboSyncHandler(s),
		Trade:           NewNuboTradeHandler(s),
		User:            NewNuboUserHandler(s),
	}
}

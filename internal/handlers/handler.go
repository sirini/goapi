package handlers

import "github.com/sirini/goapi/internal/services"

// 모든 핸들러들을 관리
type Handler struct {
	Auth    AuthHandler
	Board   BoardHandler
	Chat    ChatHandler
	Comment CommentHandler
	Home    HomeHandler
	Noti    NotiHandler
	OAuth2  OAuth2Handler
	User    UserHandler
}

// 모든 핸들러들을 생성
func NewHandler(s *services.Service) *Handler {
	return &Handler{
		Auth:    NewTsboardAuthHandler(s),
		Board:   NewTsboardBoardHandler(s),
		Chat:    NewTsboardChatHandler(s),
		Comment: NewTsboardCommentHandler(s),
		Home:    NewTsboardHomeHandler(s),
		Noti:    NewTsboardNotiHandler(s),
		OAuth2:  NewTsboardOAuth2Handler(s),
		User:    NewTsboardUserHandler(s),
	}
}

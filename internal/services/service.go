package services

import "github.com/sirini/goapi/internal/repositories"

// 모든 서비스들을 관리
type Service struct {
	Admin   AdminService
	Auth    AuthService
	Board   BoardService
	Blog    BlogService
	Chat    ChatService
	Comment CommentService
	Home    HomeService
	Noti    NotiService
	OAuth   OAuthService
	Sync    SyncService
	Trade   TradeService
	User    UserService
}

// 모든 서비스들을 생성
func NewService(repos *repositories.Repository) *Service {
	return &Service{
		Admin:   NewTsboardAdminService(repos),
		Auth:    NewTsboardAuthService(repos),
		Board:   NewTsboardBoardService(repos),
		Blog:    NewTsboardBlogService(repos),
		Chat:    NewTsboardChatService(repos),
		Comment: NewTsboardCommentService(repos),
		Home:    NewTsboardHomeService(repos),
		Noti:    NewTsboardNotiService(repos),
		OAuth:   NewTsboardOAuthService(repos),
		Sync:    NewTsboardSyncService(repos),
		Trade:   NewTsboardTradeService(repos),
		User:    NewTsboardUserService(repos),
	}
}

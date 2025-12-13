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
		Admin:   NewNuboAdminService(repos),
		Auth:    NewNuboAuthService(repos),
		Board:   NewNuboBoardService(repos),
		Blog:    NewNuboBlogService(repos),
		Chat:    NewNuboChatService(repos),
		Comment: NewNuboCommentService(repos),
		Home:    NewNuboHomeService(repos),
		Noti:    NewNuboNotiService(repos),
		OAuth:   NewNuboOAuthService(repos),
		Sync:    NewNuboSyncService(repos),
		Trade:   NewNuboTradeService(repos),
		User:    NewNuboUserService(repos),
	}
}

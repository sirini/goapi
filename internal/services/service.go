package services

import "github.com/sirini/goapi/internal/repositories"

// 모든 서비스들을 관리
type Service struct {
	Auth  AuthService
	Chat  ChatService
	Noti  NotiService
	OAuth OAuthService
	User  UserService
}

// 모든 서비스들을 생성
func NewService(repos *repositories.Repository) *Service {
	return &Service{
		Auth:  NewTsboardAuthService(repos),
		Chat:  NewTsboardChatService(repos),
		Noti:  NewTsboardNotiService(repos),
		OAuth: NewTsboardOAuthService(repos),
		User:  NewTsboardUserService(repos),
	}
}

package services

import "github.com/sirini/goapi/internal/repositories"

// 모든 서비스들을 관리
type Service struct {
	UserService  UserService
	AuthService  AuthService
	OAuthService OAuthService
}

// 모든 서비스들을 생성
func NewService(repos *repositories.Repository) *Service {
	return &Service{
		UserService:  NewTsboardUserService(repos),
		AuthService:  NewTsboardAuthService(repos),
		OAuthService: NewTsboardOAuthService(repos),
	}
}

package services

import (
	"errors"
	"fmt"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

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

func applyPointChange(repo repositories.UserRepository, param models.UpdatePointParam) error {
	if err := repo.ApplyPointChange(param); err != nil {
		if errors.Is(err, repositories.ErrInsufficientPoint) {
			return repositories.ErrInsufficientPoint
		}
		return fmt.Errorf("failed to update point balance")
	}
	return nil
}

// 모든 서비스들을 생성
func NewService(repos *repositories.Repository) *Service {
	user := NewNuboUserService(repos)
	board := NewNuboBoardService(repos)
	mailer := utils.NewResendMailer()
	transactionalMailer := newTrackedMailer(mailer, repos.MailDelivery)
	return &Service{
		Admin:   newNuboAdminService(repos, user, mailer, mailer),
		Auth:    newNuboAuthService(repos, transactionalMailer),
		Board:   board,
		Blog:    NewNuboBlogService(repos),
		Chat:    NewNuboChatService(repos),
		Comment: newNuboCommentService(repos, transactionalMailer),
		Home:    NewNuboHomeService(repos),
		Noti:    NewNuboNotiService(repos),
		OAuth:   NewNuboOAuthService(repos),
		Sync:    NewNuboSyncService(repos),
		Trade:   NewNuboTradeService(repos, board),
		User:    user,
	}
}

package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/sirini/goapi/internal/configs"
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
	Push    PushService
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
	chat := NewNuboChatService(repos)
	mailer := utils.NewResendMailer()
	transactionalMailer := newTrackedMailer(mailer, repos.MailDelivery)
	comment := newNuboCommentService(repos, transactionalMailer)
	pushSender, err := newFirebasePushSender(
		context.Background(),
		configs.Env.FirebaseProjectID,
		configs.Env.FirebaseCredentialsFile,
	)
	if err != nil {
		log.Printf("push: Firebase sender disabled: %v", err)
	}
	notifications := newNotificationPublisher(repos, pushSender)
	board.notifications = notifications
	chat.notifications = notifications
	comment.notifications = notifications
	return &Service{
		Admin:   newNuboAdminService(repos, user, mailer, mailer),
		Auth:    newNuboAuthService(repos, transactionalMailer),
		Board:   board,
		Blog:    NewNuboBlogService(repos),
		Chat:    chat,
		Comment: comment,
		Home:    NewNuboHomeService(repos),
		Noti:    &NuboNotiService{repos: repos, publisher: notifications},
		OAuth:   NewNuboOAuthService(repos),
		Push:    NewNuboPushService(repos.Push),
		Sync:    NewNuboSyncService(repos),
		Trade:   NewNuboTradeService(repos, board),
		User:    user,
	}
}

package services

import (
	"errors"
	"time"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

var ErrChatBlocked = errors.New("chat is unavailable because of a block relation")

type ChatService interface {
	GetChattingList(userUid uint, limit uint) ([]models.ChatItem, error)
	GetChattingHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]models.ChatHistory, error)
	SaveChatMessage(actionUserUid uint, targetUserUid uint, message string) uint
	MarkChatRead(actionUserUid uint, targetUserUid uint, throughUid uint) (models.ChatReadResult, error)
}

type NuboChatService struct {
	repos         *repositories.Repository
	notifications *notificationPublisher
}

// 리포지토리 묶음 주입받기
func NewNuboChatService(repos *repositories.Repository) *NuboChatService {
	return &NuboChatService{
		repos:         repos,
		notifications: newNotificationPublisher(repos, disabledPushSender{}),
	}
}

// 쪽지 목록들 가져오기
func (s *NuboChatService) GetChattingList(userUid uint, limit uint) ([]models.ChatItem, error) {
	return s.repos.Chat.LoadChatList(userUid, limit)
}

// 상대방과의 대화내용 가져오기
func (s *NuboChatService) GetChattingHistory(actionUserUid uint, targetUserUid uint, limit uint) ([]models.ChatHistory, error) {
	// 어느 한쪽이라도 상대를 차단한 경우 기존 대화를 노출하지 않는다.
	if s.hasBlockRelation(actionUserUid, targetUserUid) {
		return []models.ChatHistory{}, nil
	}
	return s.repos.Chat.LoadChatHistory(actionUserUid, targetUserUid, limit)
}

// 다른 사용자에게 쪽지 남기기
func (s *NuboChatService) SaveChatMessage(actionUserUid uint, targetUserUid uint, message string) uint {
	if s.hasBlockRelation(actionUserUid, targetUserUid) {
		return 0
	}
	insertId := s.repos.Chat.InsertNewChat(actionUserUid, targetUserUid, utils.Escape(message))
	parameter := models.InsertNotificationParam{
		ActionUserUid: actionUserUid,
		TargetUserUid: targetUserUid,
		NotiType:      models.NOTI_CHAT_MESSAGE,
		PostUid:       0,
		CommentUid:    0,
	}
	if insertId > 0 {
		s.notifications.Save(parameter, false)
	}
	return insertId
}

// 상대방에게서 받은 쪽지를 현재 화면에 표시한 마지막 지점까지 읽음 처리한다.
func (s *NuboChatService) MarkChatRead(actionUserUid uint, targetUserUid uint, throughUid uint) (models.ChatReadResult, error) {
	if s.hasBlockRelation(actionUserUid, targetUserUid) {
		return models.ChatReadResult{}, ErrChatBlocked
	}
	readAt := uint64(time.Now().UnixMilli())
	updatedCount, err := s.repos.Chat.MarkChatRead(actionUserUid, targetUserUid, throughUid, readAt)
	if err != nil {
		return models.ChatReadResult{}, err
	}
	return models.ChatReadResult{
		ThroughUid:   throughUid,
		ReadAt:       readAt,
		UpdatedCount: updatedCount,
	}, nil
}

func (s *NuboChatService) hasBlockRelation(actionUserUid uint, targetUserUid uint) bool {
	return s.repos.User.IsBannedByTarget(actionUserUid, targetUserUid) ||
		s.repos.User.IsBannedByTarget(targetUserUid, actionUserUid)
}

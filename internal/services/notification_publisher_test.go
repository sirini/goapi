package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type notificationRepoStub struct {
	repositories.NotiRepository
	inserted int
}

func (r *notificationRepoStub) IsNotiAdded(models.InsertNotificationParam) bool { return false }
func (r *notificationRepoStub) InsertNotification(models.InsertNotificationParam) {
	r.inserted++
}
func (r *notificationRepoStub) FindUserNameProfileByUid(uint) (string, string) {
	return "사진가", ""
}

type notificationPushRepoStub struct {
	repositories.PushRepository
	removed []string
}

func (r *notificationPushRepoStub) FindTokens(uint) ([]string, error) {
	return []string{"installation-id"}, nil
}
func (r *notificationPushRepoStub) RemoveDevices(ids []string) error {
	r.removed = ids
	return nil
}

type pushSenderStub struct {
	message chan PushMessage
}

func (s *pushSenderStub) Send(
	_ context.Context,
	_ []string,
	message PushMessage,
) ([]string, error) {
	s.message <- message
	return []string{"installation-id"}, nil
}

func TestNotificationPublisherStoresAndSendsPush(t *testing.T) {
	notiRepo := &notificationRepoStub{}
	pushRepo := &notificationPushRepoStub{}
	sender := &pushSenderStub{message: make(chan PushMessage, 1)}
	publisher := newNotificationPublisher(&repositories.Repository{
		Noti: notiRepo,
		Push: pushRepo,
	}, sender)

	publisher.Save(models.InsertNotificationParam{
		ActionUserUid: 3,
		TargetUserUid: 7,
		NotiType:      models.NOTI_LEAVE_COMMENT,
		PostUid:       7522,
	}, false)

	if notiRepo.inserted != 1 {
		t.Fatalf("inserted = %d, want 1", notiRepo.inserted)
	}
	select {
	case message := <-sender.message:
		if message.Data["postUid"] != "7522" || message.Body != "사진가님이 내 사진에 댓글을 남겼습니다" {
			t.Fatalf("unexpected push message: %+v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("push was not sent")
	}
}

func TestNotificationPublisherSkipsSelfNotification(t *testing.T) {
	notiRepo := &notificationRepoStub{}
	publisher := newNotificationPublisher(&repositories.Repository{Noti: notiRepo}, disabledPushSender{})
	publisher.Save(models.InsertNotificationParam{ActionUserUid: 4, TargetUserUid: 4}, false)
	if notiRepo.inserted != 0 {
		t.Fatalf("inserted = %d, want 0", notiRepo.inserted)
	}
}

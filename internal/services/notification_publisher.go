package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type notificationPublisher struct {
	repos  *repositories.Repository
	sender PushSender
}

func newNotificationPublisher(repos *repositories.Repository, sender PushSender) *notificationPublisher {
	return &notificationPublisher{repos: repos, sender: sender}
}

func (p *notificationPublisher) Save(param models.InsertNotificationParam, deduplicate bool) {
	if param.ActionUserUid == param.TargetUserUid {
		return
	}
	if deduplicate && p.repos.Noti.IsNotiAdded(param) {
		return
	}
	p.repos.Noti.InsertNotification(param)
	p.send(param)
}

func (p *notificationPublisher) send(param models.InsertNotificationParam) {
	if p.repos.Push == nil {
		return
	}
	installationIDs, err := p.repos.Push.FindTokens(param.TargetUserUid)
	if err != nil || len(installationIDs) == 0 {
		return
	}
	name, _ := p.repos.Noti.FindUserNameProfileByUid(param.ActionUserUid)
	message := PushMessage{
		Title: configs.Env.Title,
		Body:  notificationBody(utils.Unescape(name), param.NotiType),
		Data: map[string]string{
			"type":        strconv.Itoa(int(param.NotiType)),
			"postUid":     strconv.FormatUint(uint64(param.PostUid), 10),
			"fromUserUid": strconv.FormatUint(uint64(param.ActionUserUid), 10),
		},
	}

	// 알림 저장 응답을 지연시키지 않되 외부 호출이 무한정 남지 않도록 제한한다.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		invalid, sendErr := p.sender.Send(ctx, installationIDs, message)
		if len(invalid) > 0 {
			if removeErr := p.repos.Push.RemoveDevices(invalid); removeErr != nil {
				log.Printf("push: failed to remove invalid installations: %v", removeErr)
			}
		}
		if sendErr != nil {
			log.Printf("push: failed to deliver notification to user %d: %v", param.TargetUserUid, sendErr)
		}
	}()
}

func notificationBody(name string, notificationType models.Noti) string {
	if name == "" {
		name = "누군가"
	}
	action := map[models.Noti]string{
		models.NOTI_LIKE_POST:     "내 사진을 좋아합니다",
		models.NOTI_LIKE_COMMENT:  "내 댓글을 좋아합니다",
		models.NOTI_LEAVE_COMMENT: "내 사진에 댓글을 남겼습니다",
		models.NOTI_REPLY_COMMENT: "내 댓글에 답글을 남겼습니다",
		models.NOTI_CHAT_MESSAGE:  "나에게 메시지를 보냈습니다",
	}[notificationType]
	if action == "" {
		action = "새로운 활동을 남겼습니다"
	}
	return fmt.Sprintf("%s님이 %s", name, action)
}

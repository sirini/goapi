package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type NotiService interface {
	CheckedAllNoti(userUid uint)
	CheckedSingleNoti(notiUid uint, userUid uint)
	GetUserNoti(userUid uint, limit uint) ([]models.NotificationItem, error)
	SaveNewNoti(param models.InsertNotificationParam)
}

type NuboNotiService struct {
	repos     *repositories.Repository
	publisher *notificationPublisher
}

// 리포지토리 묶음 주입받기
func NewNuboNotiService(repos *repositories.Repository) *NuboNotiService {
	return &NuboNotiService{
		repos:     repos,
		publisher: newNotificationPublisher(repos, disabledPushSender{}),
	}
}

// 모든 알람 확인 처리하기
func (s *NuboNotiService) CheckedAllNoti(userUid uint) {
	s.repos.Noti.UpdateAllChecked(userUid)
}

// 지정된 알림 번호에 대한 확인 처리하기
func (s *NuboNotiService) CheckedSingleNoti(notiUid uint, userUid uint) {
	s.repos.Noti.UpdateChecked(notiUid, userUid)
}

// 사용자의 알림 내역 가져오기
func (s *NuboNotiService) GetUserNoti(userUid uint, limit uint) ([]models.NotificationItem, error) {
	return s.repos.Noti.FindNotificationByUserUid(userUid, limit)
}

// 새로운 알림 저장하기
func (s *NuboNotiService) SaveNewNoti(param models.InsertNotificationParam) {
	s.publisher.Save(param, true)
}

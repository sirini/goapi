package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type NotiService interface {
	CheckedAllNoti(userUid uint)
	CheckedSingleNoti(notiUid uint)
	GetUserNoti(userUid uint, limit uint) ([]models.NotificationItem, error)
	SaveNewNoti(param models.InsertNotificationParameter)
}

type TsboardNotiService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardNotiService(repos *repositories.Repository) *TsboardNotiService {
	return &TsboardNotiService{repos: repos}
}

// 모든 알람 확인 처리하기
func (s *TsboardNotiService) CheckedAllNoti(userUid uint) {
	s.repos.Noti.UpdateAllChecked(userUid)
}

// 지정된 알림 번호에 대한 확인 처리하기
func (s *TsboardNotiService) CheckedSingleNoti(notiUid uint) {
	s.repos.Noti.UpdateChecked(notiUid)
}

// 사용자의 알림 내역 가져오기
func (s *TsboardNotiService) GetUserNoti(userUid uint, limit uint) ([]models.NotificationItem, error) {
	return s.repos.Noti.FindNotificationByUserUid(userUid, limit)
}

// 새로운 알림 저장하기
func (s *TsboardNotiService) SaveNewNoti(param models.InsertNotificationParameter) {
	isDup := s.repos.Noti.IsNotiAdded(param)
	if isDup || param.ActionUserUid == param.TargetUserUid {
		return
	}
	s.repos.Noti.InsertNotification(param)
}

package services

import (
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type NotiService interface {
	CheckedAllNoti(userUid uint, limit uint)
	GetUserNoti(userUid uint, limit uint) ([]models.NotificationItem, error)
	SaveNewNoti(param models.NewNotiParameter)
}

type TsboardNotiService struct {
	repos *repositories.Repository
}

// 리포지토리 묶음 주입받기
func NewTsboardNotiService(repos *repositories.Repository) *TsboardNotiService {
	return &TsboardNotiService{repos: repos}
}

// 모든 알람 확인 처리하기
func (s *TsboardNotiService) CheckedAllNoti(userUid uint, limit uint) {
	s.repos.Noti.UpdateAllChecked(userUid, limit)
}

// 사용자의 알림 내역 가져오기
func (s *TsboardNotiService) GetUserNoti(userUid uint, limit uint) ([]models.NotificationItem, error) {
	return s.repos.Noti.LoadNotification(userUid, limit)
}

// 새로운 알림 저장하기
func (s *TsboardNotiService) SaveNewNoti(param models.NewNotiParameter) {
	isDup := s.repos.Noti.IsNotiAdded(param)
	if isDup || param.ActionUserUid == param.TargetUserUid {
		return
	}
	s.repos.Noti.InsertNewNotification(param)
}

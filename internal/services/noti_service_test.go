package services

import (
	"testing"

	"github.com/sirini/goapi/internal/repositories"
)

type notificationOwnershipRepo struct {
	repositories.NotiRepository
	notiUid uint
	userUid uint
}

func (r *notificationOwnershipRepo) UpdateChecked(notiUid uint, userUid uint) {
	r.notiUid = notiUid
	r.userUid = userUid
}

func TestCheckedSingleNotificationScopesUpdateToUser(t *testing.T) {
	noti := &notificationOwnershipRepo{}
	s := NewNuboNotiService(&repositories.Repository{Noti: noti})
	s.CheckedSingleNoti(11, 22)

	if noti.notiUid != 11 || noti.userUid != 22 {
		t.Fatalf("UpdateChecked() got notification %d and user %d", noti.notiUid, noti.userUid)
	}
}

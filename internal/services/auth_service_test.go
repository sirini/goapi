package services

import (
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type authenticationStateRepo struct {
	repositories.AuthRepository
	users map[uint]models.MyInfoResult
}

func (r authenticationStateRepo) FindMyInfoByUid(userUid uint) models.MyInfoResult {
	return r.users[userUid]
}

func TestCanAuthenticateRejectsBlockedAndMissingUsers(t *testing.T) {
	auth := authenticationStateRepo{users: map[uint]models.MyInfoResult{
		1: {UserInfoResult: models.UserInfoResult{Uid: 1}},
		2: {UserInfoResult: models.UserInfoResult{Uid: 2, Blocked: true}},
	}}
	s := NewNuboAuthService(&repositories.Repository{Auth: auth})

	if !s.CanAuthenticate(1) {
		t.Fatal("active user was rejected")
	}
	if s.CanAuthenticate(2) {
		t.Fatal("blocked user was accepted")
	}
	if s.CanAuthenticate(3) {
		t.Fatal("missing user was accepted")
	}
}

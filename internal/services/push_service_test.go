package services

import (
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

type pushRepoStub struct {
	userUid  uint
	token    string
	platform string
	removed  bool
}

func (r *pushRepoStub) SaveDevice(userUid uint, token string, platform string) error {
	r.userUid, r.token, r.platform = userUid, token, platform
	return nil
}

func (r *pushRepoStub) RemoveDevice(userUid uint, token string) error {
	r.userUid, r.token, r.removed = userUid, token, true
	return nil
}

func (r *pushRepoStub) FindTokens(uint) ([]string, error) { return nil, nil }

func TestRegisterPushDeviceNormalizesAndroidPlatform(t *testing.T) {
	repo := &pushRepoStub{}
	service := NewNuboPushService(repo)
	token := "abcdefghijklmnopqrstuvwxyz123456"

	if err := service.RegisterDevice(7, models.PushDeviceParam{Token: " " + token + " ", Platform: "ANDROID"}); err != nil {
		t.Fatal(err)
	}
	if repo.userUid != 7 || repo.token != token || repo.platform != "android" {
		t.Fatalf("saved device = (%d, %q, %q)", repo.userUid, repo.token, repo.platform)
	}
}

func TestRegisterPushDeviceRejectsInvalidInput(t *testing.T) {
	service := NewNuboPushService(&pushRepoStub{})
	for _, param := range []models.PushDeviceParam{
		{Token: "short", Platform: "android"},
		{Token: "abcdefghijklmnopqrstuvwxyz123456", Platform: "ios"},
	} {
		if err := service.RegisterDevice(7, param); err == nil {
			t.Fatalf("invalid device was accepted: %+v", param)
		}
	}
}

package services

import (
	"errors"
	"mime/multipart"
	"testing"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type failingUserUpdateRepo struct {
	repositories.UserRepository
}

func (failingUserUpdateRepo) UpdateUserInfoString(uint, string, string) error {
	return errors.New("database update failed")
}

func TestChangeUserInfoPropagatesRepositoryError(t *testing.T) {
	s := NewNuboUserService(&repositories.Repository{User: failingUserUpdateRepo{}})
	err := s.ChangeUserInfo(models.UpdateUserInfoParam{UserUid: 1, Name: "name"})
	if err == nil {
		t.Fatal("user information update error was ignored")
	}
}

func TestChangeUserProfileReturnsFileOpenError(t *testing.T) {
	s := NewNuboUserService(&repositories.Repository{})
	profile := &multipart.FileHeader{Filename: "missing-profile.png", Size: 1}
	if err := s.ChangeUserProfile(1, profile, ""); err == nil {
		t.Fatal("profile file open error was ignored")
	}
}

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

type accountUserRepoStub struct {
	repositories.UserRepository
	blocked   uint
	unblocked uint
	deleted   uint
}

func (r *accountUserRepoStub) InsertBlackList(_ uint, target uint) error {
	r.blocked = target
	return nil
}

func (r *accountUserRepoStub) RemoveBlackList(_ uint, target uint) error {
	r.unblocked = target
	return nil
}

func (r *accountUserRepoStub) DeleteAccount(userUid uint) ([]string, error) {
	r.deleted = userUid
	return nil, nil
}

type accountAuthRepoStub struct {
	repositories.AuthRepository
}

func (accountAuthRepoStub) FindMyInfoByUid(uint) models.MyInfoResult {
	return models.MyInfoResult{}
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

func TestBlockAndUnblockUserRejectSelfAndUpdateRepository(t *testing.T) {
	repo := &accountUserRepoStub{}
	service := NewNuboUserService(&repositories.Repository{User: repo})
	if err := service.BlockUser(7, 7); err == nil {
		t.Fatal("self block was accepted")
	}
	if err := service.BlockUser(7, 9); err != nil || repo.blocked != 9 {
		t.Fatalf("block result = (%d, %v)", repo.blocked, err)
	}
	if err := service.UnblockUser(7, 9); err != nil || repo.unblocked != 9 {
		t.Fatalf("unblock result = (%d, %v)", repo.unblocked, err)
	}
}

func TestDeleteAccountRequiresExplicitConfirmation(t *testing.T) {
	repo := &accountUserRepoStub{}
	service := NewNuboUserService(&repositories.Repository{
		Auth: accountAuthRepoStub{},
		User: repo,
	})
	if err := service.DeleteAccount(7, "delete"); err == nil {
		t.Fatal("weak deletion confirmation was accepted")
	}
	if err := service.DeleteAccount(7, "DELETE"); err != nil || repo.deleted != 7 {
		t.Fatalf("delete result = (%d, %v)", repo.deleted, err)
	}
}

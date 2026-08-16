package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirini/goapi/internal/configs"
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

type transactionalAuthRepo struct {
	repositories.AuthRepository
	userUID  uint
	recent   bool
	savedUID uint
	saved    bool
	deleted  uint
}

func (r *transactionalAuthRepo) FindUserUidById(string) uint { return r.userUID }
func (r *transactionalAuthRepo) VerificationRecentlyIssued(string, time.Duration) bool {
	return r.recent
}
func (r *transactionalAuthRepo) SaveVerificationCode(string, string) uint {
	r.saved = true
	return r.savedUID
}
func (r *transactionalAuthRepo) DeleteVerificationCode(uid uint) { r.deleted = uid }

type transactionalUserRepo struct {
	repositories.UserRepository
}

func (transactionalUserRepo) IsEmailDuplicated(string) bool      { return false }
func (transactionalUserRepo) IsNameDuplicated(string, uint) bool { return false }

type recordingMailer struct {
	configured bool
	message    models.MailMessage
	err        error
}

func (m *recordingMailer) Configured() bool { return m.configured }
func (m *recordingMailer) Status() models.MailStatus {
	return models.MailStatus{Configured: m.configured, Provider: "resend"}
}
func (m *recordingMailer) Send(message models.MailMessage) (models.MailDelivery, error) {
	m.message = message
	return models.MailDelivery{Provider: "resend", MessageID: "email_test"}, m.err
}

func withMailConfig(t *testing.T) {
	t.Helper()
	previous := configs.Env
	configs.Env.Title = "NUBO Test"
	configs.Env.Domain = "https://community.example.com"
	t.Cleanup(func() { configs.Env = previous })
}

func TestSignupRequiresConfiguredMailer(t *testing.T) {
	repo := &transactionalAuthRepo{savedUID: 42}
	mailer := &recordingMailer{configured: false}
	service := newNuboAuthService(&repositories.Repository{
		Auth: repo,
		User: transactionalUserRepo{},
	}, mailer)

	_, err := service.Signup(models.SignupParam{ID: "member@example.com", Name: "member"})
	if err != ErrMailNotConfigured {
		t.Fatalf("Signup() error = %v", err)
	}
	if repo.saved || mailer.message.To != "" {
		t.Fatal("unconfigured signup attempted delivery")
	}
}

func TestSignupRendersTrustedServerTemplate(t *testing.T) {
	withMailConfig(t)
	repo := &transactionalAuthRepo{savedUID: 42}
	mailer := &recordingMailer{configured: true}
	service := newNuboAuthService(&repositories.Repository{
		Auth: repo,
		User: transactionalUserRepo{},
	}, mailer)

	result, err := service.Signup(models.SignupParam{ID: "member@example.com", Name: "member"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Target != 42 || mailer.message.To != "member@example.com" {
		t.Fatalf("result/message = %#v / %#v", result, mailer.message)
	}
	if !strings.Contains(mailer.message.HTML, "NUBO Test") || !strings.Contains(mailer.message.Text, "이메일 주소를 확인") {
		t.Fatal("server template content is missing")
	}
	if !strings.HasPrefix(mailer.message.IdempotencyKey, "signup-verification/42/") {
		t.Fatalf("idempotency key = %q", mailer.message.IdempotencyKey)
	}
}

func TestResetPasswordDoesNotRevealUnknownAddress(t *testing.T) {
	mailer := &recordingMailer{configured: true}
	service := newNuboAuthService(&repositories.Repository{
		Auth: &transactionalAuthRepo{userUID: 0},
	}, mailer)

	if err := service.ResetPassword(models.ResetPasswordParam{Email: "missing@example.com"}); err != nil {
		t.Fatal(err)
	}
	if mailer.message.To != "" {
		t.Fatal("reset email sent for an unknown address")
	}
}

func TestResetPasswordDeletesCodeAfterDeliveryFailure(t *testing.T) {
	withMailConfig(t)
	repo := &transactionalAuthRepo{userUID: 7, savedUID: 91}
	mailer := &recordingMailer{configured: true, err: fmt.Errorf("provider down")}
	service := newNuboAuthService(&repositories.Repository{Auth: repo}, mailer)

	if err := service.ResetPassword(models.ResetPasswordParam{Email: "member@example.com"}); err == nil {
		t.Fatal("ResetPassword() succeeded despite delivery failure")
	}
	if repo.deleted != 91 {
		t.Fatalf("deleted verification uid = %d", repo.deleted)
	}
}

func TestGeneratedVerificationCodeIsSixDigits(t *testing.T) {
	for range 20 {
		code, err := generateVerificationCode()
		if err != nil {
			t.Fatal(err)
		}
		if len(code) != 6 || strings.Trim(code, "0123456789") != "" {
			t.Fatalf("invalid verification code %q", code)
		}
	}
}

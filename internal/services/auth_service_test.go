package services

import (
	"errors"
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
	userUID        uint
	recent         bool
	savedUID       uint
	saved          bool
	deleted        uint
	consumed       bool
	consumedEmail  string
	verificationID string
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
func (r *transactionalAuthRepo) ConsumeVerificationCode(_ uint, _ string, expectedEmail string) (string, bool) {
	r.consumedEmail = expectedEmail
	if r.consumed {
		return "", false
	}
	r.consumed = true
	return r.verificationID, true
}

type transactionalUserRepo struct {
	repositories.UserRepository
	inserted int
}

func (transactionalUserRepo) IsEmailDuplicated(string) bool      { return false }
func (transactionalUserRepo) IsNameDuplicated(string, uint) bool { return false }
func (r *transactionalUserRepo) InsertNewUser(string, string, string) uint {
	r.inserted++
	return 7
}

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
		User: &transactionalUserRepo{},
	}, mailer)

	_, err := service.Signup(models.SignupParam{ID: "member@example.com", Name: "member", Password: "Password!1"})
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
		User: &transactionalUserRepo{},
	}, mailer)

	result, err := service.Signup(models.SignupParam{ID: "member@example.com", Name: "member", Password: "Password!1"})
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

func TestSignupModesBlockUnapprovedRegistration(t *testing.T) {
	previous := configs.Env
	t.Cleanup(func() { configs.Env = previous })
	service := newNuboAuthService(&repositories.Repository{User: &transactionalUserRepo{}}, &recordingMailer{configured: true})
	param := models.SignupParam{ID: "member@example.com", Name: "member", Password: "Password!1"}

	configs.Env.SignupMode = "disabled"
	if _, err := service.Signup(param); !errors.Is(err, ErrSignupDisabled) {
		t.Fatalf("disabled Signup() error = %v", err)
	}

	configs.Env.SignupMode = "invite_only"
	if _, err := service.Signup(param); !errors.Is(err, ErrInvalidInvite) {
		t.Fatalf("invite-only Signup() error = %v", err)
	}
	if service.CanRegisterOAuthUser() {
		t.Fatal("invite-only mode allowed a new OAuth account")
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

func TestVerifyEmailBindsAddressAndConsumesCodeOnce(t *testing.T) {
	previous := configs.Env
	configs.Env.SignupMode = "verified_email"
	t.Cleanup(func() { configs.Env = previous })
	authRepo := &transactionalAuthRepo{verificationID: "member@example.com"}
	userRepo := &transactionalUserRepo{}
	service := newNuboAuthService(&repositories.Repository{Auth: authRepo, User: userRepo}, &recordingMailer{})
	param := models.VerifyParam{
		Target: 42, Code: "123456", ID: "member@example.com", Password: "Password!1", Name: "member",
	}

	ok, err := service.VerifyEmail(param)
	if err != nil || !ok {
		t.Fatalf("first VerifyEmail() = %v, %v", ok, err)
	}
	if authRepo.consumedEmail != param.ID || userRepo.inserted != 1 {
		t.Fatalf("verification email/inserts = %q / %d", authRepo.consumedEmail, userRepo.inserted)
	}
	ok, err = service.VerifyEmail(param)
	if err != nil || ok || userRepo.inserted != 1 {
		t.Fatalf("replayed VerifyEmail() = %v, %v; inserts = %d", ok, err, userRepo.inserted)
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

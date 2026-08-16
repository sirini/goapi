package services

import (
	"errors"
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

type recordingDeliveryRepo struct {
	record models.MailDeliveryRecord
	err    error
}

func (r *recordingDeliveryRepo) CreateDelivery(record models.MailDeliveryRecord) error {
	r.record = record
	return r.err
}

func (r *recordingDeliveryRepo) ListDeliveries(param models.MailDeliveryListParam, since uint64) (models.MailDeliveryListResult, error) {
	return models.MailDeliveryListResult{}, nil
}

func TestTrackedMailerRecordsAcceptedTransactionalDelivery(t *testing.T) {
	repo := &recordingDeliveryRepo{}
	mailer := &recordingMailer{configured: true}
	tracked := newTrackedMailer(mailer, repo)

	delivery, err := tracked.Send(models.MailMessage{
		To:      "member@example.com",
		Subject: "Welcome",
		HTML:    "<p>secret content</p>",
		Text:    "secret content",
		Tags:    map[string]string{"type": "signup-verification"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if delivery.MessageID != "email_test" {
		t.Fatalf("delivery = %#v", delivery)
	}
	if repo.record.Status != models.MailDeliveryAccepted || repo.record.Type != "signup-verification" {
		t.Fatalf("record = %#v", repo.record)
	}
	if repo.record.ProviderMessageID != "email_test" || repo.record.Recipient != "member@example.com" {
		t.Fatalf("record = %#v", repo.record)
	}
	if repo.record.Created == 0 {
		t.Fatal("delivery timestamp was not recorded")
	}
}

func TestTrackedMailerRecordsProviderFailureWithoutMaskingIt(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	repo := &recordingDeliveryRepo{}
	tracked := newTrackedMailer(&recordingMailer{configured: true, err: providerErr}, repo)

	_, err := tracked.Send(models.MailMessage{
		To:      "member@example.com",
		Subject: "Comment notification",
		HTML:    "<p>comment</p>",
		Tags:    map[string]string{"type": "comment-notification"},
	})
	if !errors.Is(err, providerErr) {
		t.Fatalf("Send() error = %v", err)
	}
	if repo.record.Status != models.MailDeliveryFailed || repo.record.Error != providerErr.Error() {
		t.Fatalf("record = %#v", repo.record)
	}
}

func TestTrackedMailerHistoryFailureDoesNotTurnAcceptedMailIntoFailure(t *testing.T) {
	repo := &recordingDeliveryRepo{err: errors.New("database unavailable")}
	tracked := newTrackedMailer(&recordingMailer{configured: true}, repo)

	if _, err := tracked.Send(models.MailMessage{To: "member@example.com", Subject: "Notice", HTML: "<p>notice</p>"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if repo.record.Type != "transactional" {
		t.Fatalf("record type = %q", repo.record.Type)
	}
}

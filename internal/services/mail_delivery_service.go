package services

import (
	"log"
	"strings"
	"time"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type trackedMailer struct {
	mailer utils.Mailer
	repo   repositories.MailDeliveryRepository
}

func newTrackedMailer(mailer utils.Mailer, repo repositories.MailDeliveryRepository) *trackedMailer {
	return &trackedMailer{mailer: mailer, repo: repo}
}

func (m *trackedMailer) Configured() bool {
	return m.mailer.Configured()
}

func (m *trackedMailer) Status() models.MailStatus {
	return m.mailer.Status()
}

func (m *trackedMailer) Send(message models.MailMessage) (models.MailDelivery, error) {
	delivery, err := m.mailer.Send(message)
	record := models.MailDeliveryRecord{
		Type:              mailDeliveryType(message.Tags),
		Recipient:         strings.TrimSpace(message.To),
		Subject:           strings.TrimSpace(message.Subject),
		Provider:          delivery.Provider,
		ProviderMessageID: delivery.MessageID,
		Status:            models.MailDeliveryAccepted,
		Created:           uint64(time.Now().UnixMilli()),
	}
	if err != nil {
		record.Status = models.MailDeliveryFailed
		record.Error = truncateMailDeliveryError(err.Error())
	}
	if m.repo != nil {
		if recordErr := m.repo.CreateDelivery(record); recordErr != nil {
			log.Printf("mail: failed to save transactional delivery history: %v", recordErr)
		}
	}
	return delivery, err
}

func mailDeliveryType(tags map[string]string) string {
	kind := strings.TrimSpace(tags["type"])
	if kind == "" {
		return "transactional"
	}
	runes := []rune(kind)
	if len(runes) > 50 {
		return string(runes[:50])
	}
	return kind
}

func truncateMailDeliveryError(message string) string {
	runes := []rune(strings.TrimSpace(message))
	if len(runes) > 1000 {
		return string(runes[:1000])
	}
	return string(runes)
}

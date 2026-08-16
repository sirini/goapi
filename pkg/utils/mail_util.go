package utils

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

const (
	resendProvider              = "resend"
	mailSendTimeout             = 10 * time.Second
	resendFreeDaily             = 100
	resendFreeMonthly           = 3000
	resendFreeMarketingContacts = 1000
)

// Mailer isolates transactional delivery from the services that compose mail.
type Mailer interface {
	Configured() bool
	Status() models.MailStatus
	Send(message models.MailMessage) (models.MailDelivery, error)
}

type ResendMailer struct {
	client    *resend.Client
	apiKey    string
	fromEmail string
	fromName  string
	replyTo   string
}

func NewResendMailer() *ResendMailer {
	fromEmail := strings.TrimSpace(configs.Env.ResendFromEmail)
	if fromEmail == "" {
		fromEmail = defaultResendFromEmail(configs.Env.Domain)
	}
	fromName := strings.TrimSpace(configs.Env.ResendFromName)
	if fromName == "" {
		fromName = strings.TrimSpace(configs.Env.Title)
	}

	return &ResendMailer{
		client:    resend.NewClient(configs.Env.ResendKey),
		apiKey:    strings.TrimSpace(configs.Env.ResendKey),
		fromEmail: fromEmail,
		fromName:  fromName,
		replyTo:   strings.TrimSpace(configs.Env.ResendReplyToEmail),
	}
}

func (m *ResendMailer) Configured() bool {
	if !strings.HasPrefix(m.apiKey, "re_") {
		return false
	}
	if m.replyTo != "" {
		address, err := mail.ParseAddress(m.replyTo)
		if err != nil || address.Address != m.replyTo {
			return false
		}
	}
	address, err := mail.ParseAddress(m.fromEmail)
	return err == nil && address.Address == m.fromEmail
}

func (m *ResendMailer) Status() models.MailStatus {
	from := m.fromEmail
	if m.fromName != "" && from != "" {
		from = (&mail.Address{Name: m.fromName, Address: from}).String()
	}
	status := models.MailStatus{
		Configured:            m.Configured(),
		Provider:              resendProvider,
		From:                  from,
		ReplyTo:               m.replyTo,
		DomainStatus:          "not_configured",
		FreeDaily:             resendFreeDaily,
		FreeMonthly:           resendFreeMonthly,
		FreeMarketingContacts: resendFreeMarketingContacts,
	}
	if !status.Configured {
		return status
	}

	status.DomainStatus = "unknown"
	_, domain, ok := strings.Cut(m.fromEmail, "@")
	if !ok || domain == "" {
		return status
	}
	ctx, cancel := context.WithTimeout(context.Background(), mailSendTimeout)
	defer cancel()
	domains, err := m.client.Domains.ListWithContext(ctx)
	if err != nil {
		return status
	}
	status.DomainStatus = "not_found"
	for _, item := range domains.Data {
		if strings.EqualFold(item.Name, domain) {
			status.DomainStatus = item.Status
			break
		}
	}
	return status
}

func (m *ResendMailer) Send(message models.MailMessage) (models.MailDelivery, error) {
	delivery := models.MailDelivery{Provider: resendProvider}
	if !m.Configured() {
		return delivery, fmt.Errorf("resend mail delivery is not configured")
	}
	if _, err := mail.ParseAddress(message.To); err != nil {
		return delivery, fmt.Errorf("invalid recipient address: %w", err)
	}
	if strings.TrimSpace(message.Subject) == "" || strings.TrimSpace(message.HTML) == "" {
		return delivery, fmt.Errorf("mail subject and HTML body are required")
	}

	tags := make([]resend.Tag, 0, len(message.Tags))
	keys := make([]string, 0, len(message.Tags))
	for key := range message.Tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		tags = append(tags, resend.Tag{Name: key, Value: message.Tags[key]})
	}

	ctx, cancel := context.WithTimeout(context.Background(), mailSendTimeout)
	defer cancel()
	response, err := m.client.Emails.SendWithOptions(ctx, &resend.SendEmailRequest{
		From:    (&mail.Address{Name: m.fromName, Address: m.fromEmail}).String(),
		To:      []string{message.To},
		Subject: message.Subject,
		Html:    message.HTML,
		Text:    message.Text,
		ReplyTo: m.replyTo,
		Tags:    tags,
	}, &resend.SendEmailOptions{IdempotencyKey: message.IdempotencyKey})
	if err != nil {
		return delivery, fmt.Errorf("resend rejected the email: %w", err)
	}
	if response == nil || response.Id == "" {
		return delivery, fmt.Errorf("resend returned no message id")
	}
	delivery.MessageID = response.Id
	return delivery, nil
}

func defaultResendFromEmail(domain string) string {
	parsed, err := url.Parse(strings.TrimSpace(domain))
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	host := parsed.Hostname()
	if host == "localhost" {
		return ""
	}
	return "noreply@" + host
}

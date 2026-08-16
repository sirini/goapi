package models

// MailMessage is a provider-neutral transactional email request.
type MailMessage struct {
	To             string
	Subject        string
	HTML           string
	Text           string
	IdempotencyKey string
	Tags           map[string]string
}

// MailDelivery identifies an email accepted by the configured provider.
type MailDelivery struct {
	Provider  string
	MessageID string
}

// MailStatus exposes configuration readiness without revealing credentials.
type MailStatus struct {
	Configured            bool   `json:"configured"`
	Provider              string `json:"provider"`
	From                  string `json:"from"`
	ReplyTo               string `json:"replyTo"`
	DomainStatus          string `json:"domainStatus"`
	FreeDaily             uint   `json:"freeDaily"`
	FreeMonthly           uint   `json:"freeMonthly"`
	FreeMarketingContacts uint   `json:"freeMarketingContacts"`
}

const (
	MailDeliveryAccepted = "accepted"
	MailDeliveryFailed   = "failed"
)

// MailDeliveryRecord is NUBO's provider-independent audit record for a
// transactional delivery attempt. Message bodies and verification secrets are
// deliberately excluded.
type MailDeliveryRecord struct {
	Uid               uint   `json:"uid"`
	Type              string `json:"type"`
	Recipient         string `json:"recipient"`
	Subject           string `json:"subject"`
	Provider          string `json:"provider"`
	ProviderMessageID string `json:"providerMessageId"`
	Status            string `json:"status"`
	Error             string `json:"error"`
	Created           uint64 `json:"created"`
}

type MailDeliveryListParam struct {
	Page  uint
	Limit uint
}

type MailDeliverySummary struct {
	Since               uint64 `json:"since"`
	Accepted            uint   `json:"accepted"`
	Failed              uint   `json:"failed"`
	SignupVerification  uint   `json:"signupVerification"`
	PasswordReset       uint   `json:"passwordReset"`
	CommentNotification uint   `json:"commentNotification"`
}

type MailDeliveryListResult struct {
	Items   []MailDeliveryRecord `json:"items"`
	Total   uint                 `json:"total"`
	Page    uint                 `json:"page"`
	Limit   uint                 `json:"limit"`
	Summary MailDeliverySummary  `json:"summary"`
}

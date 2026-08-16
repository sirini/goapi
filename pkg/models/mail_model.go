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

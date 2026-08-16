package models

const (
	MailCampaignDraft   = "draft"
	MailCampaignSyncing = "syncing"
	MailCampaignReady   = "ready"
	MailCampaignSending = "sending"
	MailCampaignSent    = "sent"
	MailCampaignFailed  = "failed"
)

type MailCampaignSaveParam struct {
	Uid      uint   `json:"uid"`
	Subject  string `json:"subject"`
	Markdown string `json:"markdown"`
}

type MailCampaignUidParam struct {
	Uid uint `json:"uid"`
}

type MailCampaignPreviewParam struct {
	Subject  string `json:"subject"`
	Markdown string `json:"markdown"`
}

type MailCampaign struct {
	Uid               uint   `json:"uid"`
	Subject           string `json:"subject"`
	Markdown          string `json:"markdown"`
	Status            string `json:"status"`
	RecipientCount    uint   `json:"recipientCount"`
	ResendSegmentId   string `json:"-"`
	ResendImportId    string `json:"-"`
	ResendBroadcastId string `json:"resendBroadcastId"`
	LastError         string `json:"lastError"`
	Created           uint64 `json:"created"`
	Updated           uint64 `json:"updated"`
	Sent              uint64 `json:"sent"`
}

type MailCampaignListResult struct {
	Items []MailCampaign `json:"items"`
	Total uint           `json:"total"`
}

type MailCampaignPreviewResult struct {
	HTML string `json:"html"`
	Text string `json:"text"`
}

type MailRecipient struct {
	Email string
	Name  string
}

type MailImportStatus struct {
	Status  string
	Total   uint
	Failed  uint
	Skipped uint
}

type MarketingBroadcast struct {
	SegmentId string
	Name      string
	Subject   string
	HTML      string
	Text      string
}

package models

const (
	BADGE_FIRST_POST    = "first-post"
	BADGE_FIRST_COMMENT = "first-comment"
	BADGE_SENSTA_APP    = "sensta-app"

	CLIENT_SENSTA_ANDROID = "sensta-android"
	CLIENT_SENSTA_IOS     = "sensta-ios"
	CLIENT_HEADER         = "X-Nubo-Client"
	APP_VERSION_HEADER    = "X-Nubo-App-Version"
)

func IsSenstaClient(clientKey string) bool {
	return clientKey == CLIENT_SENSTA_ANDROID || clientKey == CLIENT_SENSTA_IOS
}

// UserBadge is an achievement a user keeps after qualifying for it once.
type UserBadge struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconKey     string `json:"iconKey"`
	EarnedAt    uint64 `json:"earnedAt"`
}

// BadgeDefinition describes a permanent achievement that can be awarded to users.
// System definitions are rule-backed and cannot be edited from the administrator UI.
type BadgeDefinition struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconKey     string `json:"iconKey"`
	Active      bool   `json:"active"`
	ShowInline  bool   `json:"showInline"`
	SortOrder   uint   `json:"sortOrder"`
	System      bool   `json:"system"`
	Created     uint64 `json:"created"`
	Updated     uint64 `json:"updated"`
}

type AdminBadgeDefinitionParam struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconKey     string `json:"iconKey"`
	ShowInline  bool   `json:"showInline"`
	SortOrder   uint   `json:"sortOrder"`
}

type AdminBadgeGrantParam struct {
	UserUid  uint   `json:"userUid"`
	BadgeKey string `json:"badgeKey"`
}

type BadgeAcknowledgeParam struct {
	Keys []string `json:"keys"`
}

// BadgeAwardParam records the evidence for an idempotent achievement grant.
// GrantedBy is zero for automatic grants and the acting administrator UID for manual grants.
type BadgeAwardParam struct {
	UserUid      uint
	BadgeKey     string
	QualifiedAt  uint64
	GrantSource  string
	GrantedBy    uint
	EvidenceType string
	EvidenceUid  uint
}

type PostOriginParam struct {
	PostUid    uint
	ClientKey  string
	AppVersion string
	RecordedAt uint64
}

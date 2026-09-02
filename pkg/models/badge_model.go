package models

const (
	BADGE_FIRST_POST    = "first-post"
	BADGE_FIRST_COMMENT = "first-comment"
	BADGE_SENSTA_APP    = "sensta-app"

	CLIENT_SENSTA_ANDROID = "sensta-android"
	CLIENT_HEADER         = "X-Nubo-Client"
	APP_VERSION_HEADER    = "X-Nubo-App-Version"
)

// UserBadge is an achievement a user keeps after qualifying for it once.
type UserBadge struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconKey     string `json:"iconKey"`
	EarnedAt    uint64 `json:"earnedAt"`
}

// BadgeAwardParam records the evidence for an idempotent achievement grant.
// GrantedBy is zero for automatic grants and reserved for a future administrator grant UI.
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

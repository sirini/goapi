package models

// 새 알림 추가 파라미터 정의
type NewNotiParameter struct {
	ActionUserUid uint
	TargetUserUid uint
	NotiType      uint
	PostUid       uint
	CommentUid    uint
}

// 알림내용 조회 항목 정의
type NotificationItem struct {
	Uid       uint          `json:"uid"`
	FromUser  UserBasicInfo `json:"fromUser"`
	Type      Noti          `json:"type"`
	Id        string        `json:"id"`
	BoardType uint          `json:"boardType"`
	PostUid   uint          `json:"postUid"`
	Checked   bool          `json:"checked"`
	Timestamp uint64        `json:"timestamp"`
}

// 알림 타입 재정의
type Noti uint8

// 알림 타입 고유값들
const (
	NOTI_LIKE_POST = iota
	NOTI_LIKE_COMMENT
	NOTI_LEAVE_COMMENT
	NOTI_REPLY_COMMENT
	NOTI_CHAT_MESSAGE
)

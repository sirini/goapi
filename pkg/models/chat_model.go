package models

// 쪽지 목록용 정보
type ChatItem struct {
	Sender    UserBasicInfo `json:"sender"`
	Uid       uint          `json:"uid"`
	Message   string        `json:"message"`
	Timestamp uint64        `json:"timestamp"`
}

// 쪽지 내용 보기용 정보
type ChatHistory struct {
	Uid       uint   `json:"uid"`
	UserUid   uint   `json:"userUid"`
	Message   string `json:"message"`
	Timestamp uint64 `json:"timestamp"`
	ReadAt    uint64 `json:"readAt"`
}

// 쪽지 보내기에 필요한 파라미터 정의
type ChatSendMessage struct {
	TargetUserUid uint   `json:"targetUserUid"`
	Message       string `json:"message"`
}

// 쪽지를 실제로 표시한 마지막 지점까지 읽음 처리하는 파라미터
type ChatReadRequest struct {
	TargetUserUid uint `json:"targetUserUid"`
	ThroughUid    uint `json:"throughUid"`
}

// 읽음 처리 결과
type ChatReadResult struct {
	ThroughUid   uint   `json:"throughUid"`
	ReadAt       uint64 `json:"readAt"`
	UpdatedCount int64  `json:"updatedCount"`
}

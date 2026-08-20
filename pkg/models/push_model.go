package models

// 모바일 푸시 토큰 등록·해제 요청
type PushDeviceParam struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

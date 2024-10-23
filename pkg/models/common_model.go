package models

// 가장 기본적인 서버 응답
type ResponseCommon struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Result  interface{} `json:"result"`
}

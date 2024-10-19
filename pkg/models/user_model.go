package models

// 사용자 정보
type UserInfo struct {
	Uid       uint   `json:"uid"`
	Name      string `json:"name"`
	Profile   string `json:"profile"`
	Level     uint   `json:"level"`
	Signature string `json:"signature"`
	Signup    uint64 `json:"signup"`
	Signin    uint64 `json:"signin"`
	Admin     bool   `json:"admin"`
	Blocked   bool   `json:"blocked"`
}

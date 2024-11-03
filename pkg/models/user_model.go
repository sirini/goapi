package models

import "mime/multipart"

// (공개된) 사용자 정보
type UserInfoResult struct {
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

// (로그인 한) 내 정보
type MyInfoResult struct {
	UserInfoResult
	Id      string `json:"id"`
	Point   uint   `json:"point"`
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

// 권한 확인이 필요한 액션 정의
type Action uint8

// 액션 고유 값들
const (
	WRITE_POST = iota
	WRITE_COMMENT
	SEND_CHAT
	SEND_REPORT
)

// 액션 이름 반환
func (a Action) String() string {
	switch a {
	case WRITE_COMMENT:
		return "write_comment"
	case SEND_CHAT:
		return "send_chat"
	case SEND_REPORT:
		return "send_report"
	default:
		return "write_post"
	}
}

// 내 정보 수정하기 파라미터 정의
type UpdateUserInfoParameter struct {
	UserUid        uint
	Name           string
	Signature      string
	Password       string
	Profile        multipart.File
	ProfileHandler *multipart.FileHeader
}

// 사용자의 권한 정보들
type UserPermissionResult struct {
	WritePost       bool `json:"writePost"`
	WriteComment    bool `json:"writeComment"`
	SendChatMessage bool `json:"sendChatMessage"`
	SendReport      bool `json:"sendReport"`
}

// 사용자 권한 및 로그인, 신고 내역 정의
type UserPermissionReportResult struct {
	UserPermissionResult
	Login    bool   `json:"login"`
	UserUid  uint   `json:"userUid"`
	Response string `json:"response"`
}

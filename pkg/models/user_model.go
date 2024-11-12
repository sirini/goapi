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

// 액션 타입 재정의
type UserAction uint8

// 액션 고유 값들
const (
	ACTION_WRITE_POST UserAction = iota
	ACTION_WRITE_COMMENT
	ACTION_SEND_CHAT
	ACTION_SEND_REPORT
)

// 액션 이름 반환
func (a UserAction) String() string {
	switch a {
	case ACTION_WRITE_COMMENT:
		return "write_comment"
	case ACTION_SEND_CHAT:
		return "send_chat"
	case ACTION_SEND_REPORT:
		return "send_report"
	default:
		return "write_post"
	}
}

// 사용자 포인트 변경 이력 타입 정의
type PointAction uint

// 포인트 변경 액션들
const (
	POINT_ACTION_VIEW PointAction = iota
	POINT_ACTION_WRITE
	POINT_ACTION_COMMENT
	POINT_ACTION_DOWNLOAD
)

// 포인트 변경 액션 이름 반환
func (pa PointAction) String() string {
	switch pa {
	case POINT_ACTION_WRITE:
		return "write"
	case POINT_ACTION_COMMENT:
		return "comment"
	case POINT_ACTION_DOWNLOAD:
		return "download"
	default:
		return "view"
	}
}

// 포인트 변경 파라미터 정의
type UpdatePointParameter struct {
	UserUid  uint
	BoardUid uint
	Action   PointAction
	Point    uint
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

// 사용자의 최소 기본 정보들
type UserBasicInfo struct {
	UserUid uint   `json:"uid"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

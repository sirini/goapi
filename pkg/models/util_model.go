package models

// 업로드 폴더의 하위 폴더들
type UploadCategory string

// 하위 폴더들의 상수 정의
const (
	UPLOAD_ATTACH  UploadCategory = "attachments"
	UPLOAD_IMAGE   UploadCategory = "images"
	UPLOAD_PROFILE UploadCategory = "profile"
	UPLOAD_TEMP    UploadCategory = "temp"
	UPLOAD_THUMB   UploadCategory = "thumbnails"
)

// uid, id, name 3개를 담는 구조체
type Triple struct {
	Pair
	Id string `json:"id"`
}

// 에러 코드 정의
type Code uint

// 에러 타입 상수 정의
const (
	CODE_SUCCESS Code = iota
	CODE_NOT_ADMIN
	CODE_INVALID_TOKEN
	CODE_INVALID_PARAMETER
	CODE_FAILED_OPERATION
	CODE_DUPLICATED_VALUE
	CODE_NO_PERMISSION
	CODE_EXCEED_SIZE
	CODE_EXPIRED_TOKEN
)
package models

// 업로드 폴더의 하위 폴더들
type UploadCategory string

// 하위 폴더들의 상수 정의
const (
	ATTACH  UploadCategory = "attachments"
	PROFILE UploadCategory = "profile"
	TEMP    UploadCategory = "temp"
	THUMB   UploadCategory = "thumbnails"
)

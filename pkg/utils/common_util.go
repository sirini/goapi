package utils

import (
	"html/template"
	"time"

	"github.com/sirini/goapi/pkg/models"
)

// 참이면 1, 거짓이면 0 반환
func ToUint(b bool) uint {
	if b {
		return 1
	}
	return 0
}

// HTML 문자열을 이스케이프
func Escape(raw string) string {
	return template.HTMLEscapeString(raw)
}

// YYYY:mm:dd HH:ii:ss 형태의 시간 문자를 Unix timestamp로 변경
func ConvUnixMilli(timeStr string) uint64 {
	layout := "2006:01:02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return models.FAILED
	}
	return uint64(t.UnixMilli())
}

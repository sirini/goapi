package utils

import (
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/sirini/goapi/pkg/models"
)

// HTML 문자열을 이스케이프
func Escape(raw string) string {
	safeStr := template.HTMLEscapeString(raw)
	safeStr = strings.ReplaceAll(safeStr, "&#34;", "&quot;")
	safeStr = strings.ReplaceAll(safeStr, "&#39;", "&#x27;")
	return safeStr
}

// HTML 문자열 이스케이프 해제
func Unescape(escaped string) string {
	originStr := html.UnescapeString(escaped)
	originStr = strings.ReplaceAll(originStr, "&quot;", "\"")
	originStr = strings.ReplaceAll(originStr, "&#x27;", "'")
	return originStr
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

// Unix timestamp 형식의 숫자를 YYYY:mm:dd HH:ii:ss 형태로 변경
func ConvTimestamp(timestamp uint64) string {
	t := time.UnixMilli(int64(timestamp))
	return t.Format("2006:01:02 15:04:05")
}

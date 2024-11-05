package utils

import "html/template"

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

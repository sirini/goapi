package utils

import (
	"regexp"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	once           sync.Once
	sanitizePolicy *bluemonday.Policy
)

// int 절대값 구하기
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Sanitize 정책 초기화
func initSanitizePolicy() {
	sanitizePolicy = bluemonday.NewPolicy()
	sanitizePolicy.AllowElements(
		"h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "p", "a",
		"ul", "ol", "nl", "li", "b", "i", "strong", "em",
		"strike", "code", "hr", "br", "div", "table",
		"thead", "caption", "tbody", "tr", "th", "td", "pre", "img",
	)
	sanitizePolicy.AllowAttrs("href", "name", "target").OnElements("a")
	sanitizePolicy.AllowAttrs("src", "alt").OnElements("img")
}

// Sanitize 정책 가져오기
func getSanitizePolicy() *bluemonday.Policy {
	once.Do(initSanitizePolicy)
	return sanitizePolicy
}

// 입력 문자열 중 HTML 태그들은 허용된 것만 남겨두기
func Sanitize(input string) string {
	policy := getSanitizePolicy()
	return policy.Sanitize(input)
}

// 순수한 문자(영어는 소문자), 숫자만 남기고 특수기호, 공백 등은 제거
func Purify(input string) string {
	re := regexp.MustCompile(`[^\p{L}\d]`)
	result := re.ReplaceAllString(input, "")
	return strings.ToLower(result)
}

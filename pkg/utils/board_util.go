package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/microcosm-cc/bluemonday"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
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

// 글 작성/수정 시 파라미터 검사 및 타입 변환
func CheckWriteParameters(c fiber.Ctx) (models.EditorWriteParameter, error) {
	result := models.EditorWriteParameter{}
	actionUserUid := ExtractUserUid(c.Get("Authorization"))
	boardUid, err := strconv.ParseUint(c.FormValue("boardUid"), 10, 32)
	if err != nil {
		return result, err
	}
	categoryUid, err := strconv.ParseUint(c.FormValue("categoryUid"), 10, 32)
	if err != nil {
		return result, err
	}
	isNotice, err := strconv.ParseBool(c.FormValue("isNotice"))
	if err != nil {
		return result, err
	}
	isSecret, err := strconv.ParseBool(c.FormValue("isSecret"))
	if err != nil {
		return result, err
	}

	title := Escape(c.FormValue("title"))
	if len(title) < 2 {
		return result, fmt.Errorf("invalid title, too short")
	}
	title = CutString(title, 299)

	content := Sanitize(c.FormValue("content"))
	if len(content) < 2 {
		return result, fmt.Errorf("invalid content, too short")
	}

	tags := c.FormValue("tags")
	tagArr := strings.Split(tags, ",")

	fileSizeLimit, _ := strconv.ParseInt(configs.Env.FileSizeLimit, 10, 32)
	form, err := c.MultipartForm()
	if err != nil {
		return result, err
	}
	attachments := form.File["attachments[]"]
	if len(attachments) > 0 {
		var totalFileSize int64
		for _, fileHeader := range attachments {
			totalFileSize += fileHeader.Size
		}

		if totalFileSize > fileSizeLimit {
			return result, fmt.Errorf("uploaded files exceed size limitation")
		}
	}

	result = models.EditorWriteParameter{
		BoardUid:    uint(boardUid),
		UserUid:     uint(actionUserUid),
		CategoryUid: uint(categoryUid),
		Title:       title,
		Content:     content,
		Files:       attachments,
		Tags:        tagArr,
		IsNotice:    isNotice,
		IsSecret:    isSecret,
	}
	return result, nil
}

// (한글 포함) 문자열 안전하게 자르기
func CutString(s string, max int) string {
	runeCount := 0
	for i := range s {
		runeCount++
		if runeCount > max {
			return s[:i]
		}
	}
	return s
}

// Sanitize 정책 초기화
func initSanitizePolicy() {
	sanitizePolicy = bluemonday.NewPolicy()
	sanitizePolicy.AllowElements(
		"h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "p", "a",
		"ul", "ol", "nl", "li", "b", "i", "strong", "em", "mark", "span",
		"strike", "code", "hr", "br", "div", "table",
		"thead", "caption", "tbody", "tr", "th", "td", "pre", "img",
	)
	sanitizePolicy.AllowAttrs("href", "name", "target", "style", "class").OnElements("a")
	sanitizePolicy.AllowAttrs("src", "alt", "style", "class").OnElements("img")
	sanitizePolicy.AllowAttrs("style", "class").OnElements("span")
	sanitizePolicy.AllowAttrs("class").OnElements("code")
}

// 게시글/댓글 상태값 반환
func GetContentStatus(isNotice bool, isSecret bool) models.Status {
	status := models.CONTENT_NORMAL
	if isNotice {
		status = models.CONTENT_NOTICE
	} else if isSecret {
		status = models.CONTENT_SECRET
	}
	return status
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

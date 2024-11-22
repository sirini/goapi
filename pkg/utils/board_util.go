package utils

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
func CheckWriteParameters(r *http.Request) (models.EditorWriteParameter, error) {
	result := models.EditorWriteParameter{}
	actionUserUid := GetUserUidFromToken(r)
	boardUid, err := strconv.ParseUint(r.FormValue("boardUid"), 10, 32)
	if err != nil {
		return result, err
	}
	categoryUid, err := strconv.ParseUint(r.FormValue("categoryUid"), 10, 32)
	if err != nil {
		return result, err
	}
	isNotice, err := strconv.ParseBool(r.FormValue("isNotice"))
	if err != nil {
		return result, err
	}
	isSecret, err := strconv.ParseBool(r.FormValue("isSecret"))
	if err != nil {
		return result, err
	}
	title := Escape(r.FormValue("title"))
	if len(title) < 2 {
		return result, fmt.Errorf("invalid title, too short")
	}
	content := Sanitize(r.FormValue("content"))
	if len(content) < 2 {
		return result, fmt.Errorf("invalid content, too short")
	}
	tags := r.FormValue("tags")
	tagArr := strings.Split(tags, ",")

	fileSizeLimit, _ := strconv.ParseInt(configs.Env.FileSizeLimit, 10, 32)
	err = r.ParseMultipartForm(fileSizeLimit)
	if err != nil {
		return result, err
	}
	attachments := r.MultipartForm.File["attachments"]
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
		UserUid:     actionUserUid,
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

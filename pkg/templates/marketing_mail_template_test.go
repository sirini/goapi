package templates

import (
	"strings"
	"testing"
)

func TestRenderMarketingMailSanitizesMarkdownAndAddsUnsubscribe(t *testing.T) {
	html, text, err := RenderMarketingMail("NUBO", "https://example.com", "소식", "# 안녕하세요\n\n<script>alert(1)</script>\n\n[링크](javascript:alert(1))", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script") || strings.Contains(html, "javascript:") {
		t.Fatalf("unsafe marketing HTML: %s", html)
	}
	if !strings.Contains(html, unsubscribePlaceholder) || !strings.Contains(text, unsubscribePlaceholder) {
		t.Fatal("unsubscribe placeholder is missing")
	}
}

func TestRenderMarketingTestMailOmitsUnsubscribePlaceholder(t *testing.T) {
	html, text, err := RenderMarketingMail("NUBO", "https://example.com", "테스트", "본문", true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, unsubscribePlaceholder) || strings.Contains(text, unsubscribePlaceholder) {
		t.Fatal("test mail contains a Resend unsubscribe placeholder")
	}
}

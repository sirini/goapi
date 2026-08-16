package templates

import (
	"strings"
	"testing"
)

func TestRenderTransactionalMailEscapesDynamicContent(t *testing.T) {
	html, text, err := RenderTransactionalMail(MailContent{
		SiteName:  "NUBO",
		SiteURL:   "https://example.com",
		Heading:   "새 댓글",
		Body:      `<script>alert("x")</script>`,
		Highlight: `<img src=x onerror=alert(1)>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") || strings.Contains(html, "<img src=x") {
		t.Fatalf("dynamic HTML was not escaped: %s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") || !strings.Contains(text, `<script>alert("x")</script>`) {
		t.Fatalf("expected escaped HTML and readable plain text")
	}
}

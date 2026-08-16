package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/sirini/goapi/pkg/utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const unsubscribePlaceholder = "{{{RESEND_UNSUBSCRIBE_URL}}}"
const unsubscribeSentinel = "https://nubo.invalid/unsubscribe"

type marketingMailContent struct {
	SiteName    string
	SiteURL     string
	Subject     string
	Body        template.HTML
	Test        bool
	Unsubscribe string
}

var marketingMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

var marketingMailTemplate = template.Must(template.New("marketing-mail").Parse(`<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light">
  <title>{{.Subject}}</title>
  <style>
    .nubo-content h1,.nubo-content h2,.nubo-content h3{margin:30px 0 12px;color:#2f2a25;line-height:1.35;letter-spacing:-.025em}
    .nubo-content h1{font-size:26px}.nubo-content h2{font-size:22px}.nubo-content h3{font-size:18px}
    .nubo-content p,.nubo-content li{font-size:16px;line-height:1.8;color:#514941}
    .nubo-content p{margin:0 0 18px}.nubo-content ul,.nubo-content ol{margin:0 0 20px;padding-left:24px}
    .nubo-content a{color:#9a4f2e}.nubo-content blockquote{margin:22px 0;padding:4px 0 4px 18px;border-left:3px solid #cfaa8d;color:#6f6257}
    .nubo-content code{padding:2px 6px;border-radius:5px;background:#f3eee7;font-family:ui-monospace,monospace;font-size:.9em}
    .nubo-content pre{overflow:auto;margin:22px 0;padding:18px;border-radius:10px;background:#302b27;color:#fffaf3}
    .nubo-content pre code{padding:0;background:transparent;color:inherit}.nubo-content img{max-width:100%;height:auto;border-radius:10px}
    .nubo-content hr{margin:30px 0;border:0;border-top:1px solid #e5ddd2}.nubo-content table{width:100%;border-collapse:collapse;margin:22px 0}
    .nubo-content th,.nubo-content td{padding:10px;border:1px solid #ded7cc;text-align:left}.nubo-content th{background:#f5efe7}
  </style>
</head>
<body style="margin:0;padding:0;background:#f3efe7;color:#332f2a;font-family:-apple-system,BlinkMacSystemFont,'Apple SD Gothic Neo','Malgun Gothic','Segoe UI',sans-serif;">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">{{.Subject}}</div>
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width:100%;background:#f3efe7;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width:100%;max-width:660px;">
        <tr><td style="padding:0 8px 20px;font-size:20px;font-weight:700;color:#3a332c;"><a href="{{.SiteURL}}" style="color:#3a332c;text-decoration:none;">{{.SiteName}}</a></td></tr>
        <tr><td style="background:#fffdf9;border:1px solid #ded7cc;border-radius:16px;padding:42px 40px;box-shadow:0 12px 30px rgba(67,55,43,.06);">
          <p style="margin:0 0 14px;font-size:12px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;color:#a0522d;">Newsletter</p>
          <h1 style="margin:0 0 28px;font-size:28px;line-height:1.3;letter-spacing:-.03em;color:#2f2a25;">{{.Subject}}</h1>
          <div class="nubo-content">{{.Body}}</div>
        </td></tr>
        <tr><td style="padding:22px 8px 0;font-size:12px;line-height:1.8;color:#877d72;">
          이 메일은 <a href="{{.SiteURL}}" style="color:#9a4f2e;text-decoration:none;">{{.SiteName}}</a> 회원에게 발송되었습니다.<br>
          {{if .Test}}테스트 메일에서는 수신 거부 링크가 활성화되지 않습니다.{{else}}더 이상 단체 메일을 받고 싶지 않다면 <a href="{{.Unsubscribe}}" style="color:#9a4f2e;">수신을 거부할 수 있습니다.</a>{{end}}
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`))

func RenderMarketingMail(siteName, siteURL, subject, markdown string, test bool) (string, string, error) {
	var rendered bytes.Buffer
	if err := marketingMarkdown.Convert([]byte(markdown), &rendered); err != nil {
		return "", "", err
	}
	policy := bluemonday.UGCPolicy()
	cleanHTML := policy.SanitizeBytes(rendered.Bytes())
	if strings.TrimSpace(utils.PlainText(string(cleanHTML))) == "" {
		return "", "", fmt.Errorf("mail content is empty")
	}

	unsubscribe := ""
	if !test {
		unsubscribe = unsubscribeSentinel
	}
	var output bytes.Buffer
	if err := marketingMailTemplate.Execute(&output, marketingMailContent{
		SiteName: siteName, SiteURL: siteURL, Subject: subject,
		Body: template.HTML(cleanHTML), Test: test, Unsubscribe: unsubscribe,
	}); err != nil {
		return "", "", err
	}
	html := output.String()
	text := strings.Join([]string{siteName, subject, utils.PlainText(string(cleanHTML))}, "\n\n")
	if !test {
		html = strings.ReplaceAll(html, unsubscribeSentinel, unsubscribePlaceholder)
		text += "\n\n수신 거부: " + unsubscribePlaceholder
	}
	return html, text, nil
}

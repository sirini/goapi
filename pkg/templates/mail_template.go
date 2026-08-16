package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

type MailContent struct {
	SiteName    string
	SiteURL     string
	Preheader   string
	Label       string
	Heading     string
	Greeting    string
	Body        string
	Highlight   string
	ActionLabel string
	ActionURL   string
	Notice      string
}

var transactionalMailTemplate = template.Must(template.New("transactional-mail").Parse(`<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light">
  <meta name="supported-color-schemes" content="light">
  <title>{{.Heading}}</title>
</head>
<body style="margin:0;padding:0;background:#f3efe7;color:#332f2a;font-family:-apple-system,BlinkMacSystemFont,'Apple SD Gothic Neo','Malgun Gothic','Segoe UI',sans-serif;">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">{{.Preheader}}</div>
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width:100%;background:#f3efe7;">
    <tr>
      <td align="center" style="padding:40px 16px;">
        <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width:100%;max-width:600px;">
          <tr>
            <td style="padding:0 8px 20px;font-size:20px;font-weight:700;letter-spacing:-0.02em;color:#3a332c;">
              <a href="{{.SiteURL}}" style="color:#3a332c;text-decoration:none;">{{.SiteName}}</a>
            </td>
          </tr>
          <tr>
            <td style="background:#fffdf9;border:1px solid #ded7cc;border-radius:16px;padding:42px 40px;box-shadow:0 12px 30px rgba(67,55,43,0.06);">
              <p style="margin:0 0 14px;font-size:12px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#a0522d;">{{.Label}}</p>
              <h1 style="margin:0 0 24px;font-size:28px;line-height:1.3;letter-spacing:-0.03em;color:#2f2a25;">{{.Heading}}</h1>
              {{if .Greeting}}<p style="margin:0 0 12px;font-size:16px;line-height:1.75;color:#514941;">{{.Greeting}}</p>{{end}}
              <p style="margin:0;font-size:16px;line-height:1.75;color:#514941;white-space:pre-line;">{{.Body}}</p>
              {{if .Highlight}}
              <div style="margin:28px 0 0;padding:22px 24px;border-radius:12px;background:#f5e8dc;border:1px solid #ead4c2;font-size:24px;font-weight:700;line-height:1.5;letter-spacing:0.08em;text-align:center;color:#7b3f24;word-break:break-word;">{{.Highlight}}</div>
              {{end}}
              {{if .ActionURL}}
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin-top:30px;">
                <tr><td style="border-radius:10px;background:#9a4f2e;"><a href="{{.ActionURL}}" style="display:inline-block;padding:14px 22px;font-size:15px;font-weight:700;color:#ffffff;text-decoration:none;">{{.ActionLabel}}</a></td></tr>
              </table>
              {{end}}
              {{if .Notice}}<p style="margin:30px 0 0;padding-top:22px;border-top:1px solid #ebe5dc;font-size:13px;line-height:1.7;color:#81776d;">{{.Notice}}</p>{{end}}
            </td>
          </tr>
          <tr>
            <td style="padding:22px 8px 0;font-size:12px;line-height:1.7;color:#877d72;">
              이 메일은 <a href="{{.SiteURL}}" style="color:#9a4f2e;text-decoration:none;">{{.SiteName}}</a>에서 발송되었습니다.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

func RenderTransactionalMail(content MailContent) (string, string, error) {
	var html bytes.Buffer
	if err := transactionalMailTemplate.Execute(&html, content); err != nil {
		return "", "", err
	}

	plainParts := []string{content.SiteName, content.Heading}
	if content.Greeting != "" {
		plainParts = append(plainParts, content.Greeting)
	}
	plainParts = append(plainParts, content.Body)
	if content.Highlight != "" {
		plainParts = append(plainParts, content.Highlight)
	}
	if content.ActionURL != "" {
		plainParts = append(plainParts, fmt.Sprintf("%s: %s", content.ActionLabel, content.ActionURL))
	}
	if content.Notice != "" {
		plainParts = append(plainParts, content.Notice)
	}
	return html.String(), strings.Join(plainParts, "\n\n"), nil
}

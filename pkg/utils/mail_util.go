package utils

import (
	"strings"

	"github.com/resend/resend-go/v2"
	"github.com/sirini/goapi/internal/configs"
)

// Resend API를 이용해서 메일 발송하기 (무료: 일 100건 / 월 3,000건 제한 있음)
func SendMailByResend(to string, from string, subject string, body string) bool {
	client := resend.NewClient(configs.Env.ResendKey)
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Html:    body,
		Subject: subject,
	}
	_, err := client.Emails.Send(params)
	return err == nil
}

// Resend API Key가 설정되어 있으면 메일 발송
func SendMail(to string, from string, subject string, body string) bool {
	if !strings.HasPrefix(configs.Env.ResendKey, "re_") {
		return false
	}
	return SendMailByResend(to, from, subject, body)
}

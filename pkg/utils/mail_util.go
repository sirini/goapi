package utils

import (
	"strings"

	"github.com/resend/resend-go/v2"
	"github.com/sirini/goapi/internal/configs"
	"gopkg.in/gomail.v2"
)

// Gmail의 앱비밀번호를 이용해서 메일 발송하기 (일 ~500회 이내 제한 있음)
func SendMailByGmail(to string, subject string, body string) bool {
	m := gomail.NewMessage()
	m.SetHeader("From", configs.Env.GmailID)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, configs.Env.GmailID, configs.Env.GmailAppPassword)
	err := d.DialAndSend(m)
	return err == nil
}

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

// Resend가 있으면 우선 이용하고, 없다면 Gmail을 차선으로 이용해서 메일 발송
func SendMail(to string, from string, subject string, body string) bool {
	if strings.HasPrefix(configs.Env.ResendKey, "re_") {
		if ok := SendMailByResend(to, from, subject, body); ok {
			return true
		}
	}
	if len(configs.Env.GmailAppPassword) == 16 {
		if ok := SendMailByGmail(to, subject, body); ok {
			return true
		}
	}
	return false
}

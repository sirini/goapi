package utils

import (
	"github.com/sirini/goapi/internal/configs"
	"gopkg.in/gomail.v2"
)

func SendMail(to string, subject string, body string) bool {
	m := gomail.NewMessage()
	m.SetHeader("From", configs.Env.GmailID)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, configs.Env.GmailID, configs.Env.GmailAppPassword)

	err := d.DialAndSend(m)
	if err != nil {
		return false
	}
	return true
}

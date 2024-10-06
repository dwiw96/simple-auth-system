package email

import (
	"fmt"
	"html/template"

	"github.com/wneessen/go-mail"
)

const CONFIG_SMTP_HOST = "live.smtp.mailtrap.io"
const CONFIG_SMTP_PORT = 587
const CONFIG_SENDER_NAME = "Simple Auth System - Go <hello@demomailtrap.com>"
const CONFIG_FROM_EMAIL = "hello@demomailtrap.com"
const CONFIG_EMAIL_USERNAME = "smtp@mailtrap.io"
const CONFIG_AUTH_PASSWORD = "afeff63ccc8241f259052ffa4174d1e2" // "pchd clak vavp akhy"

func SendEmail(to, subject, htmlPath string, urlData map[string]interface{}) (err error) {
	m := mail.NewMsg()
	if err = m.From(CONFIG_FROM_EMAIL); err != nil {
		return fmt.Errorf("failed to set mail From address, err:  %v", err)
	}
	if err = m.To(to); err != nil {
		return fmt.Errorf("failed to set mail To address, err: %v", err)
	}

	m.Subject(subject)

	tmpl, err := template.ParseFiles(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %s", err)
	}

	if err := m.SetBodyHTMLTemplate(tmpl, urlData); err != nil {
		return fmt.Errorf("failed to set HTML template mail body: %s", err)
	}

	c, err := mail.NewClient(CONFIG_SMTP_HOST, mail.WithPort(CONFIG_SMTP_PORT), mail.WithSMTPAuth(mail.SMTPAuthType(mail.SMTPAuthPlain)),
		mail.WithUsername(CONFIG_EMAIL_USERNAME), mail.WithPassword(CONFIG_AUTH_PASSWORD))
	if err != nil {
		return fmt.Errorf("failed to create mail client, err: %v", err)
	}

	if err = c.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email to %s, err: %v", to, err)
	}

	return
}

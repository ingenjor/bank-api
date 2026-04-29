package integration

import (
	"crypto/tls"

	"github.com/go-mail/mail/v2"
)

// EmailSender интерфейс для отправки писем.
type EmailSender interface {
	Send(to, subject, body string) error
}

// SMTPEmailSender реализует EmailSender через SMTP.
type SMTPEmailSender struct {
	dialer *mail.Dialer
	from   string
}

func NewEmailSender(host string, port int, user, pass string) *SMTPEmailSender {
	d := mail.NewDialer(host, port, user, pass)
	d.TLSConfig = &tls.Config{ServerName: host}
	return &SMTPEmailSender{dialer: d, from: user}
}

func (s *SMTPEmailSender) Send(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	return s.dialer.DialAndSend(m)
}

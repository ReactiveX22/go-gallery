package models

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

const (
	DefaultSender = "support@goweb.com" // TODO: make better email.
)

type Email struct {
	From      string
	To        string
	Subject   string
	Plaintext string
	HTML      string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewEmailService(config SMTPConfig) *EmailService {
	es := EmailService{
		dialer: gomail.NewDialer(config.Host, config.Port, config.Username, config.Password),
	}
	return &es
}

type EmailService struct {
	DefaultSender string

	// unexported fields
	dialer *gomail.Dialer
}

func (es *EmailService) Send(email Email) error {
	m := gomail.NewMessage()
	m.SetHeader("To", email.To)

	// set from
	es.setFrom(m, email)

	m.SetHeader("Subject", email.Subject)

	switch {
	case email.Plaintext != "" && email.HTML != "":
		m.SetBody("text/plain", email.Plaintext)
		m.AddAlternative("text/html", email.HTML)
	case email.Plaintext != "":
		m.SetBody("text/plain", email.Plaintext)
	case email.HTML != "":
		m.SetBody("text/html", email.HTML)
	}

	err := es.dialer.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (es *EmailService) ForgotPassword(to, resetURL string) error {
	email := Email{
		To:        to,
		Subject:   "Reset your password",
		Plaintext: fmt.Sprintf("Click here to reset your password: %s", resetURL),
		HTML:      fmt.Sprintf(`<a href="%s">Click here to reset your password</a>`, resetURL),
	}
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("forgot password: %w", err)
	}
	return nil
}

func (es *EmailService) setFrom(msg *gomail.Message, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	msg.SetHeader("From", from)
}

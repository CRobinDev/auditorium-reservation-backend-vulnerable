package mail

import (
	"bytes"
	"fmt"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/infra/env"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/mailtmpl"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"gopkg.in/gomail.v2"
	"html/template"
	"sync"
)

type IMailer interface {
	Send(recipientEmail, subject, templateName string, data map[string]any) error
}

type mailer struct {
	dialer    *gomail.Dialer
	templates *template.Template
}

var (
	instance IMailer
	once     sync.Once
)

func NewMailDialer() IMailer {
	once.Do(func() {
		// Parse all templates at startup
		templates, err := template.ParseFS(mailtmpl.Templates, "*.html")
		if err != nil {
			log.Fatal(map[string]interface{}{
				"error": err.Error(),
			}, "[MAIL][NewMailDialer] failed to parse templates")
			return
		}

		instance = &mailer{
			dialer: gomail.NewDialer(
				env.GetEnv().SmtpHost,
				env.GetEnv().SmtpPort,
				env.GetEnv().SmtpUsername,
				env.GetEnv().SmtpPassword,
			),
			templates: templates,
		}
	})

	return instance
}

func (m *mailer) Send(recipientEmail, subject, templateName string, data map[string]any) error {
	var tmplOutput bytes.Buffer

	err := m.templates.ExecuteTemplate(&tmplOutput, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	mail := gomail.NewMessage()
	mail.SetHeader("From", "Auditorium Reservation <"+env.GetEnv().SmtpEmail+">")
	mail.SetHeader("To", recipientEmail)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", tmplOutput.String())

	return m.dialer.DialAndSend(mail)
}

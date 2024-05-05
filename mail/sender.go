package mail

import (
	"fmt"

	"github.com/wneessen/go-mail"
)

const MailtrapSMTPHost = "sandbox.smtp.mailtrap.io"

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type MailtrapSender struct {
	name              string
	fromEmailAddress  string
	mailtrapLogin     string
	fromEmailPassword string
}

func NewMailtrapSender(name, email, login, pass string) EmailSender {
	return &MailtrapSender{
		name:              name,
		fromEmailAddress:  email,
		mailtrapLogin:     login,
		fromEmailPassword: pass,
	}
}

func (sender MailtrapSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	m := mail.NewMsg()
	if err := m.FromFormat(sender.name, sender.fromEmailAddress); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := m.To(to...); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	if err := m.Cc(cc...); err != nil {
		return fmt.Errorf("failed to set Cc address: %s", err)
	}
	if err := m.Bcc(bcc...); err != nil {
		return fmt.Errorf("failed to set Bcc address: %s", err)
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, content)
	for _, file := range attachFiles {
		m.AttachFile(file)
	}

	c, err := mail.NewClient(MailtrapSMTPHost, mail.WithPort(2525),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(sender.mailtrapLogin),
		mail.WithPassword(sender.fromEmailPassword),
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)
	}

	return c.DialAndSend(m)
}

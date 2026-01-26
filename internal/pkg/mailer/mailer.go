package mailer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/email"
)

type Mailer struct {
	mailService hermes.MailService
	config      Config
}

type Config struct {
	ClientBaseURL string
	AppName       string
	SupportEmail  string
}

func New(mailService hermes.MailService, config Config) *Mailer {
	return &Mailer{
		mailService: mailService,
		config:      config,
	}
}

type MailerInput struct {
	TemplateKey email.TemplateKey `validate:"required"`
	TargetName  string
	TargetEmail htypes.Email `validate:"required"`
	Links       []email.Link
	Languages   []string
	// Extra Params for specific templates
	NewEmail string
}

func (m Mailer) Execute(input MailerInput) error {
	for i, link := range input.Links {
		if !strings.HasPrefix(link.URL, "/") {
			continue
		}

		if strings.HasSuffix(m.config.ClientBaseURL, "path=") {
			link.URL = url.QueryEscape(link.URL)
		}

		input.Links[i].URL = m.config.ClientBaseURL + link.URL
	}

	receiver := input.TargetEmail
	t := email.Template(input.TemplateKey, input.Languages...)
	t.Execute(email.Params{
		UserName:      input.TargetName,
		ReceiverEmail: receiver.String(),
		NewEmail:      input.NewEmail,
		AppName:       m.config.AppName,
		SupportEmail:  m.config.SupportEmail,
	})

	err := m.mailService.SendCustomEmail(receiver, t.Subject.Content, t.Descriptors(input.Links...))
	if err != nil {
		return fmt.Errorf("failed to send %v email: %w", input.TemplateKey, err)
	}

	return nil
}

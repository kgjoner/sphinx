package mailer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/email"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Mail struct {
	MailService hermes.MailService
}

type MailInput struct {
	TemplateKey email.TemplateKey `validate:"required"`
	Target      auth.User
	Links       []email.Link
	Languages   []string
	// Indicates if the email is being sent to the pending email
	ToPending bool
}

func (i Mail) Execute(input MailInput) error {
	for i, link := range input.Links {
		if !strings.HasPrefix(link.URL, "/") {
			continue
		}

		if strings.HasSuffix(config.Env.CLIENT.BASE_URL, "path=") {
			link.URL = url.QueryEscape(link.URL)
		}

		input.Links[i].URL = config.Env.CLIENT.BASE_URL + link.URL
	}

	receiver := input.Target.Email
	if input.ToPending {
		receiver = input.Target.PendingEmail
	}

	t := email.Template(input.TemplateKey, input.Languages...)
	t.Execute(email.Params{
		UserName:      input.Target.Name(),
		ReceiverEmail: receiver.String(),
		NewEmail:      input.Target.PendingEmail.String(),
		AppName:       config.Env.APP_NAME,
		SupportEmail:  config.Env.SUPPORT_EMAIL,
	})

	err := i.MailService.SendCustomEmail(receiver, t.Subject.Content, t.Descriptors(input.Links...))
	if err != nil {
		return fmt.Errorf("failed to send %v email: %w", input.TemplateKey, err)
	}

	return nil
}

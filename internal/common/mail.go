package common

import (
	"net/url"
	"strings"

	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Mail struct {
	MailService hermes.MailService
}

type MailInput struct {
	TemplateKey string
	Target      auth.User
	Links       []i18n.CustomLink
	Languages   []string
	// Indicates if the email is being sent to the pending email
	ToPending bool
}

func (i Mail) Execute(input MailInput) (bool, error) {
	appName := config.Env.APP_NAME
	opt := []hermes.Options{}

	for i, link := range input.Links {
		if !strings.HasPrefix(link.Link, "/") {
			continue
		}

		if strings.HasSuffix(config.Env.CLIENT.BASE_URL, "path=") {
			link.Link = url.QueryEscape(link.Link)
		}

		input.Links[i].Link = config.Env.CLIENT.BASE_URL + link.Link
	}

	receiver := input.Target.Email
	if input.ToPending {
		receiver = input.Target.PendingEmail
	}

	t := i18n.Resource(input.Languages).Mails[input.TemplateKey]
	t.ParseContent(i18n.ResourceParams{
		UserName:      input.Target.Name(),
		ReceiverEmail: receiver.String(),
		NewEmail:      input.Target.PendingEmail.String(),
		AppName:       appName,
		SupportEmail:  config.Env.SUPPORT_EMAIL,
	})

	err := i.MailService.SendCustomEmail(receiver, t.Subject.Content, t.FormatBody(input.Links...), opt...)
	if err != nil {
		return false, err
	}

	return true, nil
}

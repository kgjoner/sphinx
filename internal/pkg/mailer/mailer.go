package mailer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/pkg/mailer/internal/assets"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Mailer struct {
	mailService hermes.MailService
	config      Config
}

type Config struct {
	ClientBaseURL        string
	AppName              string
	SupportEmail         string
	DataVerificationPath string
	PasswordResetPath    string
}

func New(mailService hermes.MailService, config Config) *Mailer {
	return &Mailer{
		mailService: mailService,
		config:      config,
	}
}

func (m *Mailer) Send(recipient htypes.Email, email shared.Email, lns ...string) error {
	links := []assets.Link{}
	params := assets.Params{
		ReceiverEmail: recipient.String(),
		AppName:       m.config.AppName,
		SupportEmail:  m.config.SupportEmail,
	}

	switch email := email.(type) {
	case identity.EmailWelcome:
		params.UserName = email.UserName
		links = append(links, assets.Link{
			Key: assets.VerificationLink,
			URL: fmt.Sprintf(
				"%v?kind=email&id=%v&code=%v",
				m.config.DataVerificationPath,
				email.UserID,
				email.Code,
			),
		})
	case identity.EmailResetPassword:
		params.UserName = email.UserName
		links = append(links, assets.Link{
			Key: assets.ResetLink,
			URL: fmt.Sprintf(
				"%v?id=%v&code=%v",
				m.config.PasswordResetPath,
				email.UserID,
				email.Code,
			),
		})
	case identity.EmailUpdateEmailNotice:
		params.UserName = email.UserName
		params.NewEmail = email.NewEmail
		links = append(links, assets.Link{
			Key: assets.CancelLink,
			URL: fmt.Sprintf(
				"%v?kind=email&action=cancel&id=%v",
				m.config.DataVerificationPath,
				email.UserID,
			),
		})
	case identity.EmailConfirmEmailUpdate:
		params.UserName = email.UserName
		params.NewEmail = email.NewEmail
		links = append(links, assets.Link{
			Key: assets.VerificationLink,
			URL: fmt.Sprintf(
				"%v?kind=email&id=%v&code=%v",
				m.config.DataVerificationPath,
				email.UserID,
				email.Code,
			),
		})
	}

	for i, link := range links {
		if !strings.HasPrefix(link.URL, "/") {
			continue
		}

		if strings.HasSuffix(m.config.ClientBaseURL, "path=") {
			link.URL = url.QueryEscape(link.URL)
		}

		links[i].URL = m.config.ClientBaseURL + link.URL
	}

	t := assets.Template(assets.TemplateKey(email.TemplateKey()), lns...)
	t.Execute(params)

	err := m.mailService.SendCustomEmail(recipient, t.Subject.Content, t.Descriptors(links...))
	if err != nil {
		return fmt.Errorf("failed to send %v email: %w", email.TemplateKey(), err)
	}

	return nil
}

func (m *Mailer) AddCustomTemplates(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}

	var templates assets.TemplateMap
	err := json.Unmarshal(raw, &templates)
	if err != nil {
		return fmt.Errorf("failed to marshal custom email templates: %w", err)
	}

	assets.MergeTemplates(templates)
	return nil
}
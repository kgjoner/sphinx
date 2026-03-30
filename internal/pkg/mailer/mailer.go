package mailer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/pkg/mailer/internal/client"
	templates "github.com/kgjoner/sphinx/internal/pkg/mailer/internal/template"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Mailer struct {
	SMTPClient       *client.Client
	Templates        templates.Templates
	FallbackLanguage string

	BaseData
	InvariantData
}

type SMTPConfig struct {
	SMTPUsername  string
	SMTPPassword  string
	SMTPHost      string
	SMTPPort      string
	AllowInsecure bool
}

func New(
	smtpConfig SMTPConfig,
	baseData BaseData,
	invariantData InvariantData,
	overrideTemplates []byte,
	defaultLanguage string,
) *Mailer {
	smtpClient := client.New(
		smtpConfig.SMTPUsername,
		smtpConfig.SMTPPassword,
		smtpConfig.SMTPHost,
		smtpConfig.SMTPPort,
		smtpConfig.AllowInsecure,
	)

	var overrides map[string]map[templates.Key]string
	if len(overrideTemplates) > 0 {
		err := json.Unmarshal(overrideTemplates, &overrides)
		if err != nil {
			panic(fmt.Errorf("failed to unmarshal custom email templates: %w", err))
		}
	}

	templs, err := templates.New(overrides)
	if err != nil {
		panic(fmt.Errorf("failed to initialize email templates: %w", err))
	}
	fmt.Println("Email templates loaded successfully", templs)

	return &Mailer{
		SMTPClient: smtpClient,
		Templates:  templs,

		BaseData:         baseData,
		InvariantData:    invariantData,
		FallbackLanguage: defaultLanguage,
	}
}

func (m *Mailer) Send(recipient prim.Email, email shared.Email, lns ...string) error {
	vData := VariantData{}

	switch email := email.(type) {
	case identity.EmailWelcome:
		vData.UserName = email.UserName
		vData.ValidationURL = m.adjustURL(
			fmt.Sprintf(
				"%v?kind=email&id=%v&code=%v",
				m.DataVerificationPath,
				email.UserID,
				email.Code,
			))
	case identity.EmailResetPassword:
		vData.UserName = email.UserName
		vData.ResetURL = m.adjustURL(
			fmt.Sprintf(
				"%v?id=%v&code=%v",
				m.PasswordResetPath,
				email.UserID,
				email.Code,
			))
	case identity.EmailUpdateEmailNotice:
		vData.UserName = email.UserName
		vData.NewEmail = email.NewEmail
		vData.CancelURL = m.adjustURL(
			fmt.Sprintf(
				"%v?kind=email&action=cancel&id=%v",
				m.DataVerificationPath,
				email.UserID,
			))
	case identity.EmailConfirmEmailUpdate:
		vData.UserName = email.UserName
		vData.NewEmail = email.NewEmail
		vData.ConfirmationURL = m.adjustURL(
			fmt.Sprintf(
				"%v?kind=email&id=%v&code=%v",
				m.DataVerificationPath,
				email.UserID,
				email.Code,
			))
	}

	var registry templates.Registry
	var exists bool
	for _, ln := range lns {
		registry, exists = m.Templates[ln]
		if exists {
			break
		}
	}

	if !exists {
		registry = m.Templates[m.FallbackLanguage]
	}

	data := data{
		BaseData:      m.BaseData,
		InvariantData: m.InvariantData,
		VariantData:   vData,
	}

	subject, body, err := registry.Execute(templates.Key(email.TemplateKey()), data)
	if err != nil {
		return err
	}

	err = m.SMTPClient.SendEmail(client.SendInput{
		To:           recipient,
		Subject:      subject,
		Body:         body,
		AliasName:    m.AliasName,
		AliasAddress: m.AliasAddress,
	})
	if err != nil {
		return fmt.Errorf("failed to send %v email: %w", email.TemplateKey(), err)
	}

	return nil
}

func (m *Mailer) UpdateData(
	baseData BaseData,
	invariantData InvariantData,
) {
	m.BaseData = baseData
	m.InvariantData = invariantData
}

func (m *Mailer) adjustURL(link string) string {
	if !strings.HasPrefix(link, "/") {
		return link
	}

	if strings.HasSuffix(m.ClientBaseURL, "path=") {
		link = url.QueryEscape(link)
	}

	return m.ClientBaseURL + link
}

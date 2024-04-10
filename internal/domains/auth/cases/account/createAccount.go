package accountcase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	"github.com/sirupsen/logrus"
)

type CreateAccount struct {
	AuthRepo    authcase.AuthRepo
	MailService hermes.MailService
}

type CreateAccountInput struct {
	auth.AccountCreationFields
	Application auth.Application
	Languages   []string
}

func (i CreateAccount) Execute(input CreateAccountInput) (*auth.Account, error) {
	acc, err := auth.NewAccount(&input.AccountCreationFields)
	if err != nil {
		return nil, err
	}

	id, err := i.AuthRepo.InsertAccount(*acc)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			pattern := regexp.MustCompile("account_(.+)_key")
			matches := pattern.FindStringSubmatch(err.Error())
			msg := fmt.Sprintf("%v has already registered", matches[1])
			return nil, normalizederr.NewRequestError(msg, "")
		}
		return nil, err
	}

	acc.InternalId = id
	err = acc.LinkTo(input.Application)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	// Send email
	t := i18n.Resource(input.Languages).Mails["welcome"];
	t.ParseContent(i18n.ResourceParams{
		UserName: acc.Name(),
	})
	
	err = i.MailService.SendCustomEmail(acc.Email, t.Subject.Content, t.FormatBody(i18n.CustomLink{
		Key: "email-verification",
		Link: fmt.Sprintf(
			"%v?kind=email&id=%v&code=%v",
			config.Environment.CLIENT_URI.DATA_VERIFICATION,
			acc.Id,
			acc.Codes[auth.AccountCodeKindValues.EMAIL_VERIFICATION],
		),
	}))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Kind":  "Mail Failed",
			"Path":  "AccountCreation",
			"Actor": acc.Id,
		}).Log(logrus.ErrorLevel, err.Error())
	}

	return acc, nil
}

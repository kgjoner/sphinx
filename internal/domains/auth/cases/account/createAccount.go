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

	t := i18n.Resource(input.Languages).Mail.Welcome
	emailVerificationLink := fmt.Sprintf(
		"%v?kind=email&id=%v&code=%v",
		config.Environment.CLIENT_URI.DATA_VERIFICATION,
		acc.Id,
		acc.Codes[auth.AccountCodeKindValues.EMAIL_VERIFICATION],
	)

	err = i.MailService.SendCustomEmail(acc.Email, fmt.Sprintf("%v %v!", t.Subject, config.Environment.APP_NAME), []hermes.CustomTemplateDescriptor{
		{
			Kind:    "title",
			Content: fmt.Sprintf("%v %v", t.Title, acc.Name()),
		},
		{
			Kind:    "text",
			Content: t.P1,
		},
		{
			Kind:    "button",
			Content: t.Button,
			Link:    emailVerificationLink,
		},
		{
			Kind:    "text",
			Content: fmt.Sprintf("%v %v", t.P2, emailVerificationLink),
		},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Kind":  "Mail Failed",
			"Path":  "AccountCreation",
			"Actor": acc.Id,
		}).Log(logrus.ErrorLevel, err.Error())
	}

	return acc, nil
}

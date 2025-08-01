package accountcase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/common/errcode"
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
	Languages []string `json:"-"`
}

func (i CreateAccount) Execute(input CreateAccountInput) (*auth.Account, error) {
	acc, err := auth.NewAccount(&input.AccountCreationFields)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationById(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Root application not found", errcode.ApplicationNotFound)
	}

	err = i.AuthRepo.InsertAccount(acc)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			pattern := regexp.MustCompile("account_(.+)_key")
			matches := pattern.FindStringSubmatch(err.Error())
			msg := fmt.Sprintf("%v has already registered", matches[1])
			return nil, normalizederr.NewConflictError(msg, errcode.DuplicateKey)
		}
		return nil, err
	}

	err = acc.GiveConsent(*app)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	// Send email
	mail := common.Mail{
		MailService: i.MailService,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "welcome",
		Target:      *acc,
		Links: []i18n.CustomLink{
			{
				Key: "email-verification",
				Link: fmt.Sprintf(
					"%v?kind=email&id=%v&code=%v",
					config.Env.CLIENT.DATA_VERIFICATION,
					acc.Id,
					acc.VerificationCodes[auth.VerificationEmail],
				),
			},
		},
		Languages: input.Languages,
	})
	if err != nil {
		i.handleError(err, *acc)
	}

	return acc, nil
}

func (i CreateAccount) handleError(err error, target auth.Account) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "AccountCreation",
		"Actor": target.Id,
	}).Log(logrus.ErrorLevel, err.Error())
}

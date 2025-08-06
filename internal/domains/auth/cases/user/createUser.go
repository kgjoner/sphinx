package usercase

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
	"github.com/sirupsen/logrus"
)

type CreateUser struct {
	AuthRepo    auth.Repo
	MailService hermes.MailService
}

type CreateUserInput struct {
	auth.UserCreationFields
	Languages []string `json:"-"`
}

func (i CreateUser) Execute(input CreateUserInput) (*auth.User, error) {
	acc, err := auth.NewUser(&input.UserCreationFields)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationByID(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Root application not found", errcode.ApplicationNotFound)
	}

	err = i.AuthRepo.InsertUser(acc)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			pattern := regexp.MustCompile("user_(.+)_key")
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
					acc.ID,
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

func (i CreateUser) handleError(err error, target auth.User) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "UserCreation",
		"Actor": target.ID,
	}).Log(logrus.ErrorLevel, err.Error())
}

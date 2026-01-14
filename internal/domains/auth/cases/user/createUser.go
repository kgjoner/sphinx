package usercase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/email"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/common/mailer"
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

func (i CreateUser) Execute(input CreateUserInput) (out auth.UserView, err error) {
	user, err := i.ExecuteEntity(input)
	if err != nil {
		return out, err
	}

	return user.View(), nil
}

// ExecuteEntity is application-internal: returns the entity for chaining.
// Only used by internal application layer (e.g., ExternalAuth).
func (i CreateUser) ExecuteEntity(input CreateUserInput) (*auth.User, error) {
	user, err := auth.NewUser(&input.UserCreationFields)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationByID(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, apperr.NewRequestError("Root application not found", errcode.ApplicationNotFound)
	}

	err = i.AuthRepo.InsertUser(user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			// TODO: rename unique constants to use "user"
			pattern := regexp.MustCompile("account_(.+)_key")
			matches := pattern.FindStringSubmatch(err.Error())
			msg := fmt.Sprintf("%v has already registered", matches[1])
			return nil, apperr.NewConflictError(msg, errcode.DuplicateKey)
		}
		return nil, err
	}

	err = user.GiveConsent(*app)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(user.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	// Send email
	mail := mailer.Mail{
		MailService: i.MailService,
	}
	err = mail.Execute(mailer.MailInput{
		TemplateKey: email.Welcome,
		Target:      *user,
		Links: []email.Link{
			{
				Key: email.VerificationLink,
				URL: fmt.Sprintf(
					"%v?kind=email&id=%v&code=%v",
					config.Env.CLIENT.DATA_VERIFICATION,
					user.ID,
					user.VerificationCodes[auth.VerificationEmail],
				),
			},
		},
		Languages: input.Languages,
	})
	if err != nil {
		i.handleError(err, *user)
	}

	return user, nil
}

func (i CreateUser) handleError(err error, target auth.User) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "UserCreation",
		"Actor": target.ID,
	}).Log(logrus.ErrorLevel, err.Error())
}

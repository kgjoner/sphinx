package usercase

import (
	"fmt"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/email"
	"github.com/kgjoner/sphinx/internal/common/mailer"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type RequestPasswordReset struct {
	AuthRepo    auth.Repo
	MailService hermes.MailService
}

type RequestPasswordResetInput struct {
	Entry     auth.Entry
	Languages []string `json:"-"`
}

func (i RequestPasswordReset) Execute(input RequestPasswordResetInput) (out bool, err error) {
	user, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, apperr.NewRequestError("User does not exist", errcode.UserNotFound)
	}

	code, err := user.RequestPasswordReset()
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	// Send email
	mail := mailer.Mail{
		MailService: i.MailService,
	}
	err = mail.Execute(mailer.MailInput{
		TemplateKey: email.PasswordReset,
		Target:      *user,
		Links: []email.Link{
			{
				Key: email.ResetLink,
				URL: fmt.Sprintf(
					"%v?id=%v&code=%v",
					config.Env.CLIENT.PASSWORD_RESET,
					user.ID,
					code,
				),
			},
		},
		Languages: input.Languages,
	})
	if err != nil {
		return out, err
	}

	return true, nil
}

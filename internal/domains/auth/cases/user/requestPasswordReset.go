package usercase

import (
	"fmt"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/common"
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

func (i RequestPasswordReset) Execute(input RequestPasswordResetInput) (bool, error) {
	user, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return false, err
	} else if user == nil {
		return false, apperr.NewRequestError("User does not exist", errcode.UserNotFound)
	}

	code, err := user.RequestPasswordReset()
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpdateUser(*user)
	if err != nil {
		return false, err
	}

	// Send email
	mail := common.Mail{
		MailService: i.MailService,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "passwordReset",
		Target:      *user,
		Links: []i18n.CustomLink{
			{
				Key: "password-reset",
				Link: fmt.Sprintf(
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
		return false, err
	}

	return true, nil
}

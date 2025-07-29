package accountcase

import (
	"fmt"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type RequestPasswordReset struct {
	AuthRepo    authcase.AuthRepo
	MailService hermes.MailService
}

type RequestPasswordResetInput struct {
	Entry       string
	Languages   []string         `json:"-"`
}

func (i RequestPasswordReset) Execute(input RequestPasswordResetInput) (bool, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry)
	if err != nil {
		return false, err
	} else if acc == nil {
		return false, normalizederr.NewRequestError("Account does not exist", errcode.AccountNotFound)
	}

	code, err := acc.RequestPasswordReset()
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpdateAccount(*acc)
	if err != nil {
		return false, err
	}

	// Send email
	mail := common.Mail{
		MailService: i.MailService,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "passwordReset",
		Target:      *acc,
		Links: []i18n.CustomLink{
			{
				Key: "password-reset",
				Link: fmt.Sprintf(
					"%v?id=%v&code=%v",
					config.Env.CLIENT.PASSWORD_RESET,
					acc.Id,
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

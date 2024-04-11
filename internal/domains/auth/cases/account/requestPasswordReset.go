package accountcase

import (
	"fmt"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/config"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type RequestPasswordReset struct {
	AuthRepo    authcase.AuthRepo
	MailService hermes.MailService
}

type RequestPasswordResetInput struct {
	Entry string
	Languages []string `json:"-"`
}

func (i RequestPasswordReset) Execute(input RequestPasswordResetInput) (bool, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry);
	if err != nil {
	 return false, err
	} else if acc == nil {
	 return false, normalizederr.NewRequestError("Account does not exist", "")
	}

	code, err := acc.RequestPasswordReset();
	if err != nil {
	 return false, err
	}

	err = i.AuthRepo.UpdateAccount(*acc);
	if err != nil {
	 return false, err
	}

	// Send email
	t := i18n.Resource(input.Languages).Mails["passwordReset"];
	t.ParseContent(i18n.ResourceParams{
		UserName: acc.Name(),
	})
	
	err = i.MailService.SendCustomEmail(acc.Email, t.Subject.Content, t.FormatBody(i18n.CustomLink{
		Key: "password-reset",
		Link: fmt.Sprintf(
			"%v?id=%v&code=%v",
			config.Environment.CLIENT_URI.PASSWORD_RESET,
			acc.Id,
			code,
		),
	}))
	if err != nil {
		return false, err
	}

	return true, nil
}

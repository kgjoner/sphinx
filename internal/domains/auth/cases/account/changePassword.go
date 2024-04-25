package accountcase

import (
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ChangePassword struct {
	AuthRepo    authcase.AuthRepo
	MailService hermes.MailService
}

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
	Languages   []string     `json:"-"`
	Actor       auth.Account `json:"-"`
}

func (i ChangePassword) Execute(input ChangePasswordInput) (bool, error) {
	acc := input.Actor
	err := acc.ChangePassword(input.OldPassword, input.NewPassword)
	if err != nil {
		return false, err
	}

	// Send email
	t := i18n.Resource(input.Languages).Mails["passwordChange"]
	t.ParseContent(i18n.ResourceParams{
		UserName: acc.Name(),
	})

	err = i.MailService.SendCustomEmail(acc.Email, t.Subject.Content, t.FormatBody())
	if err != nil {
		return false, err
	}

	// Save only after assuring notification email was sent
	err = i.AuthRepo.UpdateAccount(acc)
	if err != nil {
		return false, err
	}

	return true, nil
}

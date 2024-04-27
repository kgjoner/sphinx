package accountcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ResetPassword struct {
	AuthRepo    authcase.AuthRepo
	MailService hermes.MailService
}

type ResetPasswordInput struct {
	AccountId   uuid.UUID `json:"-"`
	Code        string
	NewPassword string
	Languages   []string `json:"-"`
}

func (i ResetPassword) Execute(input ResetPasswordInput) (bool, error) {
	acc, err := i.AuthRepo.GetAccountById(input.AccountId)
	if err != nil {
		return false, err
	} else if acc == nil {
		return false, normalizederr.NewRequestError("Account does not exist", errcode.AccountNotFound)
	}

	err = acc.ResetPassword(input.NewPassword, input.Code)
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
	err = i.AuthRepo.UpdateAccount(*acc)
	if err != nil {
		return false, err
	}
	
	err = i.AuthRepo.UpsertSessions(acc.SessionsToPersist()...)
	if err != nil {
		return false, err
	}

	return true, nil
}

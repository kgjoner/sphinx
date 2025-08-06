package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type ResetPassword struct {
	AuthRepo    auth.Repo
	MailService hermes.MailService
}

type ResetPasswordInput struct {
	UserID      uuid.UUID `json:"-"`
	Code        string
	NewPassword string
	Languages   []string `json:"-"`
}

func (i ResetPassword) Execute(input ResetPasswordInput) (bool, error) {
	acc, err := i.AuthRepo.GetUserByID(input.UserID)
	if err != nil {
		return false, err
	} else if acc == nil {
		return false, normalizederr.NewRequestError("User does not exist", errcode.UserNotFound)
	}

	err = acc.ResetPassword(input.NewPassword, input.Code)
	if err != nil {
		return false, err
	}

	// Send email
	mail := common.Mail{
		MailService: i.MailService,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "passwordChange",
		Target:      *acc,
		Languages:   input.Languages,
	})
	if err != nil {
		return false, err
	}

	// Save only after assuring notification email was sent
	err = i.AuthRepo.UpdateUser(*acc)
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpsertSessions(acc.SessionsToPersist()...)
	if err != nil {
		return false, err
	}

	return true, nil
}

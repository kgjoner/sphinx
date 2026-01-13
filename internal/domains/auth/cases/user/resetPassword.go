package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/email"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/common/mailer"
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

func (i ResetPassword) Execute(input ResetPasswordInput) (out bool, err error) {
	user, err := i.AuthRepo.GetUserByID(input.UserID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, apperr.NewRequestError("User does not exist", errcode.UserNotFound)
	}

	err = user.ResetPassword(input.NewPassword, input.Code)
	if err != nil {
		return out, err
	}

	// Send email
	mail := mailer.Mail{
		MailService: i.MailService,
	}
	err = mail.Execute(mailer.MailInput{
		TemplateKey: email.PasswordUpdateNotice,
		Target:      *user,
		Languages:   input.Languages,
	})
	if err != nil {
		return out, err
	}

	// Save only after assuring notification email was sent
	err = i.AuthRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpsertSessions(user.SessionsToPersist()...)
	if err != nil {
		return out, err
	}

	return true, nil
}

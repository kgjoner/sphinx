package usercase

import (
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/email"
	"github.com/kgjoner/sphinx/internal/common/mailer"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type ChangePassword struct {
	AuthRepo    auth.Repo
	MailService hermes.MailService
}

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
	Languages   []string  `json:"-"`
	Actor       auth.User `json:"-"`
}

func (i ChangePassword) Execute(input ChangePasswordInput) (out bool, err error) {
	user := input.Actor
	err = user.ChangePassword(input.OldPassword, input.NewPassword)
	if err != nil {
		return out, err
	}

	// Send email
	mail := mailer.Mail{
		MailService: i.MailService,
	}
	err = mail.Execute(mailer.MailInput{
		TemplateKey: email.PasswordUpdateNotice,
		Target:      user,
		Languages:   input.Languages,
	})
	if err != nil {
		return out, err
	}

	// Save only after assuring notification email was sent
	err = i.AuthRepo.UpdateUser(user)
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpsertSessions(user.SessionsToPersist()...)
	if err != nil {
		return out, err
	}

	return true, nil
}

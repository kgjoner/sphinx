package usercase

import (
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/common"
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

func (i ChangePassword) Execute(input ChangePasswordInput) (bool, error) {
	user := input.Actor
	err := user.ChangePassword(input.OldPassword, input.NewPassword)
	if err != nil {
		return false, err
	}

	// Send email
	mail := common.Mail{
		MailService: i.MailService,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "passwordChange",
		Target:      user,
		Languages:   input.Languages,
	})
	if err != nil {
		return false, err
	}

	// Save only after assuring notification email was sent
	err = i.AuthRepo.UpdateUser(user)
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpsertSessions(user.SessionsToPersist()...)
	if err != nil {
		return false, err
	}

	return true, nil
}

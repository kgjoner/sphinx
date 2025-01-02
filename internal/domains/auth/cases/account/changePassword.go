package accountcase

import (
	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ChangePassword struct {
	AuthRepo    authcase.AuthRepo
	CacheRepo   cache.DAO
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
	mail := common.Mail{
		MailService: i.MailService,
		CacheRepo: i.CacheRepo,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "passwordChange",
		Target: acc,
		Application: input.Actor.AuthedSession.Application,
		Languages: input.Languages,
	})
	if err != nil {
	 return false, err
	}

	// Save only after assuring notification email was sent
	err = i.AuthRepo.UpdateAccount(acc)
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpsertSessions(acc.SessionsToPersist()...)
	if err != nil {
		return false, err
	}

	return true, nil
}

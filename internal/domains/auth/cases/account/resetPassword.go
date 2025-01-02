package accountcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ResetPassword struct {
	AuthRepo    authcase.AuthRepo
	CacheRepo   cache.DAO
	MailService hermes.MailService
}

type ResetPasswordInput struct {
	AccountId   uuid.UUID `json:"-"`
	Code        string
	NewPassword string
	Application auth.Application
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
	mail := common.Mail{
		MailService: i.MailService,
		CacheRepo: i.CacheRepo,
	}
	_, err = mail.Execute(common.MailInput{
		TemplateKey: "passwordChange",
		Target: *acc,
		Application: input.Application,
		Languages: input.Languages,
	})
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

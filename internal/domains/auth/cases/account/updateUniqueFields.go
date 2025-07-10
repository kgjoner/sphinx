package accountcase

import (
	"fmt"

	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
	"github.com/sirupsen/logrus"
)

type UpdateUniqueFields struct {
	AuthRepo    authcase.AuthRepo
	CacheRepo   cache.DAO
	MailService hermes.MailService
}

type UpdateUniqueFieldsInput struct {
	auth.AccountUniqueFields
	Target      auth.Account `json:"-"`
	Actor       auth.Account `json:"-"`
	Application auth.Application
	Languages   []string
}

func (i UpdateUniqueFields) Execute(input UpdateUniqueFieldsInput) (*auth.AccountPrivateView, error) {
	targetAcc := &input.Target

	err := targetAcc.UpdateUniqueFields(input.AccountUniqueFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpdateAccount(*targetAcc)
	if err != nil {
		return nil, err
	}

	// Check if email is being updated
	emailBeingUpdated := !input.AccountUniqueFields.Email.IsZero()

	// Send email notice and confirmation if email was updated
	if emailBeingUpdated {
		mail := common.Mail{
			MailService: i.MailService,
			CacheRepo:   i.CacheRepo,
		}
		_, err = mail.Execute(common.MailInput{
			TemplateKey: "emailUpdateNotice",
			Target:      *targetAcc,
			Application: input.Application,
			Links: []i18n.CustomLink{
				{
					Key: "email-cancel-update",
					Link: fmt.Sprintf(
						"%v?kind=email&action=cancel&id=%v",
						config.Env.CLIENT.DATA_VERIFICATION,
						targetAcc.Id,
					),
				},
			},
			Languages: input.Languages,
		})
		if err != nil {
			i.handleError(err, *targetAcc, "notice")
		}

		secondMail := common.Mail{
			MailService: i.MailService,
			CacheRepo:   i.CacheRepo,
		}
		_, err = secondMail.Execute(common.MailInput{
			TemplateKey: "emailUpdateConfirmation",
			Target:      *targetAcc,
			ToPending:   true,
			Application: input.Application,
			Links: []i18n.CustomLink{
				{
					Key: "email-verification",
					Link: fmt.Sprintf(
						"%v?kind=email&id=%v&code=%v",
						config.Env.CLIENT.DATA_VERIFICATION,
						targetAcc.Id,
						targetAcc.Codes[auth.AccountCodeKindValues.EMAIL_VERIFICATION],
					),
				},
			},
			Languages: input.Languages,
		})
		if err != nil {
			i.handleError(err, *targetAcc, "confirmation")
		}
	}

	return targetAcc.PrivateView(input.Actor)
}

func (i UpdateUniqueFields) handleError(err error, target auth.Account, scope string) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "UpdateUniqueFields:" + scope,
		"Actor": target.Id,
	}).Log(logrus.ErrorLevel, err.Error())
}

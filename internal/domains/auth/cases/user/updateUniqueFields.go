package usercase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/common"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/sirupsen/logrus"
)

type UpdateUniqueFields struct {
	AuthRepo    auth.Repo
	MailService hermes.MailService
}

type UpdateUniqueFieldsInput struct {
	auth.UserUniqueFields
	Target    auth.User `json:"-"`
	Actor     auth.User `json:"-"`
	Languages []string  `json:"-"`
}

func (i UpdateUniqueFields) Execute(input UpdateUniqueFieldsInput) (*auth.UserPrivateView, error) {
	targetAcc := &input.Target

	if !input.Email.IsZero() && input.Email == targetAcc.Email {
		return nil, normalizederr.NewRequestError("email is already set to the same value")
	}

	if !input.Phone.IsZero() && input.Phone == targetAcc.Phone {
		return nil, normalizederr.NewRequestError("phone is already set to the same value")
	}

	err := targetAcc.UpdateUniqueFields(input.UserUniqueFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpdateUser(*targetAcc)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			pattern := regexp.MustCompile("user_(.+)_key")
			matches := pattern.FindStringSubmatch(err.Error())
			msg := fmt.Sprintf("%v has already registered", matches[1])
			return nil, normalizederr.NewRequestError(msg, errcode.DuplicateKey)
		}
		return nil, err
	}

	// Check if email is being updated
	emailBeingUpdated := !input.UserUniqueFields.Email.IsZero()

	// Send email notice and confirmation if email was updated
	if emailBeingUpdated {
		mail := common.Mail{
			MailService: i.MailService,
		}
		_, err = mail.Execute(common.MailInput{
			TemplateKey: "emailUpdateNotice",
			Target:      *targetAcc,
			Links: []i18n.CustomLink{
				{
					Key: "email-cancel-update",
					Link: fmt.Sprintf(
						"%v?kind=email&action=cancel&id=%v",
						config.Env.CLIENT.DATA_VERIFICATION,
						targetAcc.ID,
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
		}
		_, err = secondMail.Execute(common.MailInput{
			TemplateKey: "emailUpdateConfirmation",
			Target:      *targetAcc,
			ToPending:   true,
			Links: []i18n.CustomLink{
				{
					Key: "email-verification",
					Link: fmt.Sprintf(
						"%v?kind=email&id=%v&code=%v",
						config.Env.CLIENT.DATA_VERIFICATION,
						targetAcc.ID,
						targetAcc.VerificationCodes[auth.VerificationEmail],
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

func (i UpdateUniqueFields) handleError(err error, target auth.User, scope string) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "UpdateUniqueFields:" + scope,
		"Actor": target.ID,
	}).Log(logrus.ErrorLevel, err.Error())
}

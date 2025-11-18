package usercase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
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
	Field     string    `json:"-" validate:"required,oneof=email phone username document"`
	Value     string    `json:"value" validate:"required"`
	Target    auth.User `json:"-"`
	Actor     auth.User `json:"-"`
	Languages []string  `json:"-"`
}

func (i UpdateUniqueFields) Execute(input UpdateUniqueFieldsInput) (out auth.UserPrivateView, err error) {
	targetAcc := &input.Target

	uniqueFields := auth.UserUniqueFields{}
	switch input.Field {
	case "email":
		uniqueFields.Email, err = htypes.ParseEmail(input.Value)
		if err != nil {
			return out, err
		}

		if uniqueFields.Email == targetAcc.Email {
			return out, apperr.NewRequestError("email is already set to the same value")
		}
	case "phone":
		uniqueFields.Phone, err = htypes.ParsePhoneNumber(input.Value)
		if err != nil {
			return out, err
		}

		if uniqueFields.Phone == targetAcc.Phone {
			return out, apperr.NewRequestError("phone is already set to the same value")
		}
	case "username":
		uniqueFields.Username = input.Value
	case "document":
		uniqueFields.Document, err = htypes.ParseDocument(input.Value)
		if err != nil {
			return out, err
		}
	default:
		return out, apperr.NewRequestError("invalid field to update")
	}

	err = targetAcc.UpdateUniqueFields(uniqueFields)
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpdateUser(*targetAcc)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			pattern := regexp.MustCompile("user_(.+)_key")
			matches := pattern.FindStringSubmatch(err.Error())
			msg := fmt.Sprintf("%v has already registered", matches[1])
			return out, apperr.NewRequestError(msg, errcode.DuplicateKey)
		}
		return out, err
	}

	// Send email notice and confirmation if email was updated
	if input.Field == "email" {
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

package usercase

import (
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/sirupsen/logrus"
)

type UpdateUniqueField struct {
	IdentityRepo identity.Repo
	Mailer       shared.Mailer
}

type UpdateUniqueFieldInput struct {
	Field     string       `json:"-" validate:"required,oneof=email phone username document"`
	Value     string       `json:"value" validate:"required"`
	TargetID  uuid.UUID    `json:"-"`
	Actor     shared.Actor `json:"-"`
	Languages []string     `json:"-"`
}

func (i UpdateUniqueField) Execute(input UpdateUniqueFieldInput) (out identity.UserView, err error) {
	if err := identity.CanUpdateUser(&input.Actor, input.TargetID); err != nil {
		return out, err
	}

	target, err := i.IdentityRepo.GetUserByID(input.TargetID)
	if err != nil {
		return out, err
	} else if target == nil {
		return out, identity.ErrUserNotFound
	}

	switch input.Field {
	case "email":
		email, err := htypes.ParseEmail(input.Value)
		if err != nil {
			return out, err
		}

		err = target.UpdateEmail(email)
	case "phone":
		phone, err := htypes.ParsePhoneNumber(input.Value)
		if err != nil {
			return out, err
		}

		err = target.UpdatePhone(phone)
	case "username":
		if err := identity.CanUpdateUsername(&input.Actor, target); err != nil {
			return out, err
		}

		err = target.UpdateUsername(input.Value)
	case "document":
		document, err := htypes.ParseDocument(input.Value)
		if err != nil {
			return out, err
		}

		err = target.UpdateDocument(document)
	default:
		return out, identity.ErrInvalidField
	}

	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.UpdateUser(*target)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return out, identity.ErrDuplicateEntry
		}
		return out, err
	}

	// Send email notice and confirmation if email was updated
	if input.Field == "email" {
		err = i.Mailer.Send(
			target.Email,
			identity.EmailUpdateEmailNotice{
				UserName: target.Name(),
				NewEmail: target.PendingEmail.String(),
				UserID:   target.ID.String(),
			},
			input.Languages...,
		)
		if err != nil {
			i.handleError(err, *target, "notice")
		}

		err = i.Mailer.Send(
			target.PendingEmail,
			identity.EmailConfirmEmailUpdate{
				UserName: target.Name(),
				NewEmail: target.PendingEmail.String(),
				UserID:   target.ID.String(),
				Code:     target.VerificationCodes[identity.VerificationEmail],
			},
			input.Languages...,
		)
		if err != nil {
			i.handleError(err, *target, "confirmation")
		}
	}

	return target.View(), nil
}

func (i UpdateUniqueField) handleError(err error, target identity.User, scope string) {
	logrus.WithFields(logrus.Fields{
		"Kind":  "Mail Failed",
		"Path":  "UpdateUniqueField:" + scope,
		"Actor": target.ID,
	}).Log(logrus.ErrorLevel, err.Error())
}

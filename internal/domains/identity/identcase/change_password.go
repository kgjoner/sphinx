package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ChangePassword struct {
	IdentityRepo identity.Repo
	AuthRepo     auth.Repo
	Hasher       shared.PasswordHasher
	Mailer       shared.Mailer
}

type ChangePasswordInput struct {
	TargetID    uuid.UUID `json:"-"`
	OldPassword string
	NewPassword string
	Languages   []string     `json:"-"`
	Actor       shared.Actor `json:"-"`
}

func (i ChangePassword) Execute(input ChangePasswordInput) (out bool, err error) {
	if err := identity.CanUpdateUser(&input.Actor, input.TargetID); err != nil {
		return out, err
	}

	user, err := i.IdentityRepo.GetUserByID(input.TargetID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, identity.ErrUserNotFound
	}

	proof, err := shared.VerifyPassword(user.Password, input.OldPassword, i.Hasher)
	if err != nil {
		return out, err
	}

	hashPw, err := shared.NewHashedPassword(input.NewPassword, i.Hasher)
	if err != nil {
		return out, err
	}

	err = user.ChangePassword(proof, *hashPw)
	if err != nil {
		return out, err
	}

	// Send email
	err = i.Mailer.Send(
		user.Email,
		identity.EmailUpdatePasswordNotice{
			UserName: user.Name(),
		},
		input.Languages...,
	)
	if err != nil {
		return out, err
	}

	// Save only after assuring notification email was sent
	err = i.IdentityRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.TerminateAllSubjectSessions(user.ID)
	if err != nil {
		return out, err
	}

	return true, nil
}

package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ResetPassword struct {
	IdentityRepo identity.Repo
	AuthRepo     auth.Repo
	PwHasher     shared.PasswordHasher
	Mailer       shared.Mailer
}

type ResetPasswordInput struct {
	UserID      uuid.UUID `json:"-"`
	Code        string
	NewPassword string
	Languages   []string `json:"-"`
}

func (i ResetPassword) Execute(input ResetPasswordInput) (out bool, err error) {
	user, err := i.IdentityRepo.GetUserByID(input.UserID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, identity.ErrUserNotFound
	}

	proof, err := shared.VerifyCode(input.Code)
	if err != nil {
		return out, err
	}

	hashPw, err := shared.NewHashedPassword(input.NewPassword, i.PwHasher)
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

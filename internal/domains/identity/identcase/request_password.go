package identcase

import (
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type RequestPasswordReset struct {
	IdentityRepo identity.Repo
	Mailer       shared.Mailer
}

type RequestPasswordResetInput struct {
	Entry     shared.Entry
	Languages []string `json:"-"`
}

func (i RequestPasswordReset) Execute(input RequestPasswordResetInput) (out bool, err error) {
	user, err := i.IdentityRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, identity.ErrUserNotFound
	}

	code, err := user.RequestPasswordReset()
	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	// Send email
	err = i.Mailer.Send(
		user.Email,
		identity.EmailResetPassword{
			UserName: user.Name(),
			UserID:   user.ID.String(),
			Code:     code,
		},
		input.Languages...,
	)
	if err != nil {
		return out, err
	}

	return true, nil
}

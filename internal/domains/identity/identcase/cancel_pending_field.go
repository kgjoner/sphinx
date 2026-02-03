package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/identity"
)

type CancelPendingField struct {
	IdentityRepo identity.Repo
}

type CancelPendingFieldInput struct {
	UserID uuid.UUID `json:"-"`
	Field  string    `json:"-"`
}

func (i CancelPendingField) Execute(input CancelPendingFieldInput) (out bool, err error) {
	user, err := i.IdentityRepo.GetUserByID(input.UserID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, identity.ErrUserNotFound
	}

	err = user.CancelPendingField(input.Field)
	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	return true, nil
}

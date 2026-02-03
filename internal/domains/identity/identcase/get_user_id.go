package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetUserID struct {
	IdentityRepo identity.Repo
}

type GetUserIDInput struct {
	Entry shared.Entry
}

func (i GetUserID) Execute(input GetUserIDInput) (out uuid.UUID, err error) {
	user, err := i.IdentityRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return out, err
	}

	if user == nil {
		return out, nil
	}

	return user.ID, nil
}

package usercase

import (
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type CheckEntryExistence struct {
	IdentityRepo identity.Repo
}

type CheckEntryExistenceInput struct {
	Entry shared.Entry
}

func (i CheckEntryExistence) Execute(input CheckEntryExistenceInput) (out bool, err error) {
	user, err := i.IdentityRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return out, err
	}

	if user != nil {
		out = true
	}

	return out, nil
}

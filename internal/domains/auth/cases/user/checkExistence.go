package usercase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type CheckEntryExistence struct {
	AuthRepo auth.Repo
}

type CheckEntryExistenceInput struct {
	Entry auth.Entry
}

func (i CheckEntryExistence) Execute(input CheckEntryExistenceInput) (out bool, err error) {
	user, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return out, err
	}

	if user != nil {
		out = true
	}

	return out, nil
}

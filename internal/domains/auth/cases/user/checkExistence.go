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

func (i CheckEntryExistence) Execute(input CheckEntryExistenceInput) (bool, error) {
	acc, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return false, err
	}

	res := false
	if acc != nil {
		res = true
	}

	return res, nil
}

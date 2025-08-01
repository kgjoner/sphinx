package accountcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetAccountID struct {
	AuthRepo authcase.AuthRepo
}

type GetAccountIDInput struct {
	Entry auth.Entry
}

func (i GetAccountID) Execute(input GetAccountIDInput) (*uuid.UUID, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry)
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, nil
	}

	return &acc.ID, nil
}

package accountcase

import (
	"github.com/google/uuid"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetAccountId struct {
	AuthRepo authcase.AuthRepo
}

type GetAccountIdInput struct {
	Entry string
}

func (i GetAccountId) Execute(input GetAccountIdInput) (*uuid.UUID, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry);
	if err != nil {
	 return nil, err
	}

	if acc == nil {
		return nil, nil
	}

	return &acc.Id, nil
}

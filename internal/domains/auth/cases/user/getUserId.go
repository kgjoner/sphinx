package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetUserID struct {
	AuthRepo authcase.AuthRepo
}

type GetUserIDInput struct {
	Entry auth.Entry
}

func (i GetUserID) Execute(input GetUserIDInput) (*uuid.UUID, error) {
	acc, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, nil
	}

	return &acc.ID, nil
}

package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type GetUserID struct {
	AuthRepo auth.Repo
}

type GetUserIDInput struct {
	Entry auth.Entry
}

func (i GetUserID) Execute(input GetUserIDInput) (out uuid.UUID, err error) {
	user, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return out, err
	}

	if user == nil {
		return out, nil
	}

	return user.ID, nil
}

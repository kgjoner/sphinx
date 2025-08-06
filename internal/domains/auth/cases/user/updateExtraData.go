package usercase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type UpdateExtraData struct {
	AuthRepo auth.Repo
}

type UpdateExtraDataInput struct {
	auth.UserExtraFields
	Target auth.User `json:"-"`
	Actor  auth.User `json:"-"`
}

func (i UpdateExtraData) Execute(input UpdateExtraDataInput) (*auth.UserPrivateView, error) {
	targetAcc := &input.Target
	err := targetAcc.UpdateExtraData(input.UserExtraFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpdateUser(*targetAcc)
	if err != nil {
		return nil, err
	}

	return targetAcc.PrivateView(input.Actor)
}

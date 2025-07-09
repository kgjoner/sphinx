package accountcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type UpdateExtraData struct {
	AuthRepo authcase.AuthRepo
}

type UpdateExtraDataInput struct {
	ExtraData auth.ExtraData
	Target    auth.Account `json:"-"`
	Actor     auth.Account `json:"-"`
}

func (i UpdateExtraData) Execute(input UpdateExtraDataInput) (*auth.AccountPrivateView, error) {
	targetAcc := &input.Target
	err := targetAcc.UpdateExtraData(input.ExtraData)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpdateAccount(*targetAcc)
	if err != nil {
	 return nil, err
	}

	return targetAcc.PrivateView(input.Actor)
}

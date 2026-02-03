package authcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetKeyByKID struct {
	AuthRepo auth.Repo
}

type GetKeyByKIDInput struct {
	KID   string
	Actor shared.Actor
}

func (c GetKeyByKID) Execute(input GetKeyByKIDInput) (out auth.SigningKeyView, err error) {
	if err := auth.CanReadAllKeys(input.Actor); err != nil {
		return out, err
	}

	key, err := c.AuthRepo.GetSigningKeyByKID(input.KID)
	if err != nil {
		return out, err
	}

	return key.View(), nil
}

package authcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ListActiveSigningKeys struct {
	AuthRepo auth.Repo
}

type ListActiveSigningKeysInput struct {
	Actor shared.Actor
}

func (c ListActiveSigningKeys) Execute(input ListActiveSigningKeysInput) (out []auth.SigningKeyStatView, err error) {
	if err := auth.CanReadKeyStatus(input.Actor); err != nil {
		return nil, err
	}

	keys, err := c.AuthRepo.ListActiveSigningKeys()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		out = append(out, key.StatView())
	}

	return out, nil
}

func (c ListActiveSigningKeys) ExecutePublic() (out []auth.SigningKeyPubView, err error) {
	keys, err := c.AuthRepo.ListActiveSigningKeys()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		out = append(out, key.PubView())
	}

	return out, nil
}

package authcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetCurrentSigningKey struct {
	AuthRepo       auth.Repo
}

type GetCurrentSigningKeyInput struct {
	Actor shared.Actor
}

func (c GetCurrentSigningKey) Execute(input GetCurrentSigningKeyInput) (out auth.SigningKeyView, err error) {
	if err := auth.CanReadAllKeys(input.Actor); err != nil {
		return out, err
	}

	keys, err := c.AuthRepo.ListActiveSigningKeys()
	if err != nil {
		return out, err
	}

	// Find the most recent RS256 key that should be used for signing
	var currentKey *auth.SigningKey
	for _, key := range keys {
		if key.Algorithm == "RS256" && key.ShouldBeUsedForSigning() {
			if currentKey == nil || key.CreatedAt.After(currentKey.CreatedAt) {
				currentKey = key
			}
		}
	}

	if currentKey == nil {
		return out, auth.ErrNoActiveKeys
	}

	return currentKey.View(), nil
}

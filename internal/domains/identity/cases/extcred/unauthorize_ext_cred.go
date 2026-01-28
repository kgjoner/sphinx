package extcredcase

import (
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type UnauthorizeExternalCredential struct {
	IdentityRepo identity.Repo
}

type UnauthorizeExternalCredentialInput struct {
	ProviderName      string
	ProviderSubjectID string
	Actor             shared.Actor `json:"-"`
}

func (i UnauthorizeExternalCredential) Execute(input UnauthorizeExternalCredentialInput) (out bool, err error) {
	err = i.IdentityRepo.RemoveExternalCredential(input.Actor.ID, input.ProviderName, input.ProviderSubjectID)
	if err != nil {
		return out, err
	}

	return true, nil
}

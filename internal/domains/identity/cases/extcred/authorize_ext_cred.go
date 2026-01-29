package extcredcase

import (
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type AuthorizeExternalCredential struct {
	IdentityRepo     identity.Repo
	IdentityProvider shared.IdentityProvider
}

type AuthorizeExternalCredentialInput struct {
	shared.IdentityProviderInput
	Actor shared.Actor `json:"-"`
}

func (i AuthorizeExternalCredential) Execute(input AuthorizeExternalCredentialInput) (out identity.ExternalCredentialView, err error) {
	extSubject, err := i.IdentityProvider.Authenticate(input.IdentityProviderInput)
	if err != nil {
		return out, err
	} else if extSubject == nil || extSubject.ID == "" || extSubject.Email.IsZero() || extSubject.ProviderName == "" {
		return out, shared.ErrInvalidExternalSubject
	}

	user, err := i.IdentityRepo.GetUserByID(input.Actor.ID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, identity.ErrUserNotFound
	}

	extCredential := user.ExternalCredential(extSubject.ProviderName, extSubject.ID)
	if extCredential != nil {
		extCredential.SetAlias(extSubject.Alias)
		err = i.IdentityRepo.UpdateExternalCredential(*extCredential)
		if err != nil {
			return out, err
		}

		return extCredential.View(), nil
	}

	extCredential, err = user.AddExternalCredential(&identity.ExternalCredentialCreationFields{
		ProviderName:      extSubject.ProviderName,
		ProviderSubjectID: extSubject.ID,
		ProviderAlias:     extSubject.Alias,
	})
	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.InsertExternalCredential(extCredential)
	if err != nil {
		return out, err
	}

	return extCredential.View(), nil
}

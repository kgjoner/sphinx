package identcase

import (
	"github.com/kgjoner/cornucopia/v3/pwdgen"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ExternalSignUp struct {
	IdentityRepo     identity.Repo
	AccessRepo       access.Repo
	IdentityProvider shared.IdentityProvider
	PwHasher         shared.PasswordHasher
	Mailer           shared.Mailer
}

type ExternalSignUpInput struct {
	shared.IdentityProviderInput
	Languages []string `json:"-"`
}

func (i ExternalSignUp) Execute(input ExternalSignUpInput) (out identity.UserLeanView, err error) {
	extSubject, err := i.IdentityProvider.Authenticate(input.IdentityProviderInput)
	if err != nil {
		return out, err
	} else if extSubject == nil || extSubject.ID == "" || extSubject.Email.IsZero() || extSubject.ProviderName == "" {
		return out, shared.ErrInvalidExternalSubject
	}

	user, err := i.IdentityRepo.GetUserByExternalCredential(extSubject.ProviderName, extSubject.ID)
	if err != nil {
		return out, err
	}

	if user != nil {
		return out, identity.ErrExistingExternalCredential
	}

	// TODO: move pwdgen to external layer
	signUpInput := SignUpInput{
		UserCreationFields: identity.UserCreationFields{
			Email: extSubject.Email,
		},
		Password:  pwdgen.GeneratePassword(16),
		Languages: input.Languages,
	}

	userCase := SignUp{
		IdentityRepo: i.IdentityRepo,
		AccessRepo:   i.AccessRepo,
		PwHasher:     i.PwHasher,
		Mailer:       i.Mailer,
	}

	user, err = userCase.executeEntity(signUpInput)
	if err != nil {
		return out, err
	}

	extCredential, err := user.AddExternalCredential(&identity.ExternalCredentialCreationFields{
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

	return user.LeanView(), nil
}

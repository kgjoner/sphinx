package extcredcase

import (
	"github.com/kgjoner/cornucopia/v2/utils/pwdgen"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	usercase "github.com/kgjoner/sphinx/internal/domains/identity/cases/user"
	"github.com/kgjoner/sphinx/internal/shared"
)

type SignUp struct {
	IdentityRepo     identity.Repo
	AccessRepo       access.Repo
	IdentityProvider shared.IdentityProvider
	Hasher           shared.PasswordHasher
	Mailer           shared.Mailer
}

type SignUpInput struct {
	shared.IdentityProviderInput
	Languages []string `json:"-"`
}

func (i SignUp) Execute(input SignUpInput) (out identity.UserLeanView, err error) {
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
	signUpInput := usercase.SignUpInput{
		UserCreationFields: identity.UserCreationFields{
			Email: extSubject.Email,
		},
		Password:  pwdgen.GeneratePassword(16),
		Languages: input.Languages,
	}

	userCase := usercase.SignUp{
		IdentityRepo: i.IdentityRepo,
		AccessRepo:   i.AccessRepo,
		Hasher:       i.Hasher,
		Mailer:       i.Mailer,
	}

	user, err = userCase.ExecuteEntity(signUpInput)
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

	err = i.IdentityRepo.InsertUser(user)
	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.InsertExternalCredential(extCredential)
	if err != nil {
		return out, err
	}

	return user.LeanView(), nil
}

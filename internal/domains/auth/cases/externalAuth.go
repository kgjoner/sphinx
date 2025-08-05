package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/utils/pwdgen"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type ExternalAuth struct {
	AuthRepo AuthRepo
}

type ExternalAuthInput struct {
	config.ExternalAuthInput
	ProviderName string `validate:"required"`

	// Whether user consented to relate their existing account with the external provider.
	// Used only if account with the same email already exists and is not yet related.
	ConsentRelation bool

	// Whether to create a new user if no related user is found.
	// Implies user consented to relate their new account with the external provider.
	ConsentCreation bool

	// Used only when creating a new user.
	// Not required if ExternalAuthProvider already provides an email.
	Email htypes.Email

	auth.SessionCreationFields `json:"-"`
}

func (i ExternalAuth) Execute(input ExternalAuthInput) (*LoginOutput, error) {
	// Ensure actor has valid claims in external provider
	var provider *config.ExternalAuthProvider
	for _, p := range config.Env.EXTERNAL_AUTH_PROVIDERS {
		if p.Name == input.ProviderName {
			provider = &p
			break
		}
	}
	if provider == nil {
		return nil, normalizederr.NewRequestError("invalid provider", errcode.InvalidProvider)
	}

	subject, err := provider.Authenticate(input.ExternalAuthInput)
	if err != nil {
		return nil, err
	}

	// Handle user-provider relation
	user, err := i.AuthRepo.GetUserByExternalAuthID(provider.Name, subject.ID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = i.AuthRepo.GetUserByEntry(auth.Entry(subject.Email))
		if err != nil {
			return nil, err
		}

		if user != nil {
			if !input.ConsentRelation {
				return nil, normalizederr.NewForbiddenError("consent relation: an account with the same email already exists, user must consent to relate this account with the external provider.", errcode.NoConsent)
			}

			err = user.RelateToExternalProvider(provider.Name, subject.ID)
			if err != nil {
				return nil, err
			}

			err = i.AuthRepo.UpdateUser(*user)
			if err != nil {
				return nil, err
			}
		}
	}

	// Handle user creation if not found
	if user == nil {
		if !input.ConsentCreation {
			return nil, normalizederr.NewForbiddenError("consent creation: user must consent to create a new account", errcode.NoConsent)
		}

		if !input.Email.IsZero() {
			subject.Email = input.Email
		}

		if subject.Email.IsZero() {
			return nil, normalizederr.NewRequestError("there is no provided email for creating a new user", errcode.UserNotFound)
		}

		user, err = auth.NewUser(&auth.UserCreationFields{
			Email:    subject.Email,
			Password: pwdgen.Generate(16),
		})
		if err != nil {
			return nil, err
		}

		err = user.RelateToExternalProvider(provider.Name, subject.ID)
		if err != nil {
			return nil, err
		}

		err := i.AuthRepo.InsertUser(user)
		if err != nil {
			return nil, err
		}
	}

	// Handle authentication and session initialization
	err = user.AuthenticateViaExternalProvider(*subject)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationByID(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("root application not found", errcode.ApplicationNotFound)
	}
	input.Application = *app

	access, refresh, err := user.InitSession(&input.SessionCreationFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertSessions(user.SessionsToPersist()...)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		UserID:       user.ID,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}

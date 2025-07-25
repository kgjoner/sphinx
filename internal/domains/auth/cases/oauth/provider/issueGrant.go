package oauthcase

import (
	"time"

	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type IssueGrant struct {
	AuthRepo  authcase.AuthRepo
	CacheRepo cache.DAO
}

type IssueGrantInput struct {
	auth.AuthorizationGrantCreationFields
	ConsentGranted bool         `json:"consent_granted"`
	Actor          auth.Account `json:"-"`
}

func (i IssueGrant) Execute(input IssueGrantInput) (*IssueGrantOutput, error) {
	err := validator.Validate(input)
	if err != nil {
		return nil, err
	}

	// Create authorization grant
	grant, err := input.Actor.IssueAuthorizationGrant(&input.AuthorizationGrantCreationFields, input.ClientId)
	if err != nil {
		if normerr, ok := err.(normalizederr.NormalizedError); ok && normerr.Code == errcode.NoConsent {
			actorWithConsent, err := i.CreateConsentIfGranted(input)
			if err != nil {
				return nil, err
			}
			grant, err = actorWithConsent.IssueAuthorizationGrant(&input.AuthorizationGrantCreationFields, input.ClientId)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Cache the grant with TTL
	err = i.CacheRepo.CacheJson("grant:"+grant.Code, grant, time.Duration(config.Env.AUTH_GRANT_LIFETIME_IN_SEC)*time.Second)
	if err != nil {
		return nil, err
	}

	return &IssueGrantOutput{
		Code:        grant.Code,
		RedirectUri: grant.RedirectUri,
		ExpiresAt:   grant.ExpiresAt,
	}, nil
}

type IssueGrantOutput struct {
	Code        string    `json:"code"`
	RedirectUri string    `json:"redirect_uri"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (i IssueGrant) CreateConsentIfGranted(input IssueGrantInput) (*auth.Account, error) {
	if !input.ConsentGranted {
		return nil, normalizederr.NewForbiddenError("User has not consented to this application.", errcode.NoConsent)
	}

	app, err := i.AuthRepo.GetApplicationById(input.ClientId)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Application not found", errcode.ApplicationNotFound)
	}

	acc := &input.Actor
	err = acc.GiveConsent(*app)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

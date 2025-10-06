package oauthcase

import (
	"errors"
	"time"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/cornucopia/v2/repositories/cache"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type IssueGrant struct {
	AuthRepo  auth.Repo
	CacheRepo cache.DAO
}

type IssueGrantInput struct {
	auth.AuthorizationGrantCreationFields
	ConsentGranted bool      `json:"consent_granted"`
	Actor          auth.User `json:"-"`
}

func (i IssueGrant) Execute(input IssueGrantInput) (*IssueGrantOutput, error) {
	err := validator.Validate(input)
	if err != nil {
		return nil, err
	}

	// Create authorization grant
	grant, err := input.Actor.IssueAuthorizationGrant(&input.AuthorizationGrantCreationFields, input.ClientID)
	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Code == errcode.NoConsent {
			actorWithConsent, err := i.CreateConsentIfGranted(input)
			if err != nil {
				return nil, err
			}
			grant, err = actorWithConsent.IssueAuthorizationGrant(&input.AuthorizationGrantCreationFields, input.ClientID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Cache the grant with TTL
	err = i.CacheRepo.CacheJSON("grant:"+grant.Code, grant, time.Duration(config.Env.AUTH_GRANT_LIFETIME_IN_SEC)*time.Second)
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

func (i IssueGrant) CreateConsentIfGranted(input IssueGrantInput) (*auth.User, error) {
	if !input.ConsentGranted {
		return nil, apperr.NewForbiddenError("User has not consented to this application.", errcode.NoConsent)
	}

	app, err := i.AuthRepo.GetApplicationByID(input.ClientID)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, apperr.NewRequestError("Application not found", errcode.ApplicationNotFound)
	}

	user := &input.Actor
	err = user.GiveConsent(*app)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(user.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	return user, nil
}

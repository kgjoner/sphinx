package authcase

import (
	"time"

	"github.com/kgjoner/cornucopia/v2/helpers/validator"
	"github.com/kgjoner/cornucopia/v2/repositories/cache"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type IssueGrant struct {
	AuthRepo   auth.Repo
	AccessRepo access.Repo
	CacheRepo  cache.DAO
}

type IssueGrantInput struct {
	auth.GrantInput
	ConsentGranted bool         `json:"consent_granted"`
	Actor          shared.Actor `json:"-"`
}

func (i IssueGrant) Execute(input IssueGrantInput) (out IssueGrantOutput, err error) {
	targetSubID := input.Actor.ID
	targetAudID := input.ClientID
	if err = auth.CanIssueGrant(input.Actor, targetSubID, targetAudID); err != nil {
		return out, err
	}

	err = validator.Validate(input)
	if err != nil {
		return out, err
	}

	// Resolve Principal and Client
	principal, err := i.AuthRepo.GetPrincipal(targetSubID, targetAudID)
	if err != nil {
		return out, err
	}

	var client *auth.Client
	if principal == nil {
		if !input.ConsentGranted {
			return out, auth.ErrNoConsent
		}

		// TODO: remove direct reference to access application here
		app, err := i.AccessRepo.GetApplicationByID(targetAudID)
		if err != nil {
			return out, err
		} else if app == nil {
			return out, access.ErrApplicationNotFound
		}

		link, err := app.NewLink(targetSubID)
		if err != nil {
			return out, err
		}

		err = i.AccessRepo.UpsertLinks(*link)
		if err != nil {
			return out, err
		}

		roles := []string{}
		for _, r := range link.Roles {
			roles = append(roles, string(r))
		}

		principal = &auth.Principal{
			ID:         targetSubID,
			Kind:       shared.KindUser,
			Email:      input.Actor.Email,
			Name:       input.Actor.Name,
			AudienceID: targetAudID,
			Roles:      roles,
			HasConsent: true,
			IsActive:   true,
		}

		client = &auth.Client{
			ID:                  app.ID,
			Secret:              app.Secret,
			Name:                app.Name,
			AllowedRedirectUris: app.AllowedRedirectUris,
		}
	} else {
		client, err = i.AuthRepo.GetClient(input.ClientID)
		if err != nil {
			return out, err
		} else if client == nil {
			return out, auth.ErrInvalidClient
		}
	}

	proof, err := auth.VerifyConsent(*principal, *client)
	if err != nil {
		return out, err
	}

	// Create authorization grant
	grant, err := auth.NewGrant(input.GrantInput, input.Actor, *client, proof)
	if err != nil {
		return out, err
	}

	// Cache the grant with TTL
	err = i.CacheRepo.CacheJSON(
		"grant:"+grant.Code,
		grant,
		time.Duration(config.Env.AUTH_GRANT_LIFETIME_IN_SEC)*time.Second,
	)
	if err != nil {
		return out, err
	}

	return IssueGrantOutput{
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

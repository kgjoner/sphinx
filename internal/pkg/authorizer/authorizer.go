package authorizer

import (
	"context"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type authorizer struct {
	tokenProvider auth.TokenProvider
	hasher        shared.PasswordHasher
	pgPool        shared.RepoPool
	accessFactory shared.RepoFactory[access.Repo]
}

func New(tokenProvider auth.TokenProvider, hasher shared.PasswordHasher, pgPool shared.RepoPool, accessFactory shared.RepoFactory[access.Repo]) *authorizer {
	return &authorizer{
		tokenProvider: tokenProvider,
		hasher:        hasher,
		pgPool:        pgPool,
		accessFactory: accessFactory,
	}
}

func (a *authorizer) AuthorizeToken(token string, isRefreshRoute bool) (actor shared.Actor, err error) {
	sub, intent, err := a.tokenProvider.Validate(token)
	if err != nil {
		return actor, err
	}

	if intent != auth.IntentAccess && intent != auth.IntentRefresh {
		return actor, auth.ErrInvalidAccess
	}

	if (intent == auth.IntentRefresh && !isRefreshRoute) ||
		(intent == auth.IntentAccess && isRefreshRoute) {
		return actor, auth.ErrInvalidAccess
	}

	actor = shared.Actor{
		ID:          sub.ID,
		Kind:        sub.Kind,
		Email:       sub.Email,
		Name:        sub.Name,
		AudienceID:  sub.AudienceID,
		SessionID:   sub.SessionID,
		Permissions: a.mapRolesToPermissions(sub.Roles),
	}

	return actor, nil
}

func (a *authorizer) AuthorizeApp(ctx context.Context, appID uuid.UUID, appSecret string) (actor shared.Actor, err error) {
	app, err := a.accessFactory.NewDAO(ctx, a.pgPool.Connection()).GetApplicationByID(appID)
	if err != nil {
		return actor, err
	} else if app == nil {
		return actor, shared.ErrInvalidCredentials
	}

	if !a.hasher.DoesPasswordMatch(app.Secret.String(), appSecret) {
		return actor, shared.ErrInvalidCredentials
	}

	return shared.Actor{
		ID:          app.ID,
		Kind:        shared.KindApp,
		Name:        app.Name,
		AudienceID:  app.ID,
		Permissions: appPermissions,
	}, nil
}

func (a *authorizer) mapRolesToPermissions(roles []string) []string {
	permissionSet := make(map[string]struct{})
	for _, role := range roles {
		if perms, exists := permissionsByRole[role]; exists {
			for _, perm := range perms {
				permissionSet[perm] = struct{}{}
			}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions
}

package sharedhttp

import (
	"context"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Authorizer interface {
	AuthorizeToken(token string, isRefreshRoute bool) (shared.Actor, error)
	AuthorizeApp(ctx context.Context, appID uuid.UUID, appSecret string) (shared.Actor, error)
}

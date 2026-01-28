package sharedhttp

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Authorizer interface {
	AuthorizeUser(token string) (shared.Actor, auth.Intent, error)
	AuthorizeApp(appID uuid.UUID, appSecret string) (shared.Actor, error)
}

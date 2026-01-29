package identhttp

import (
	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

type gateway struct {
	Dependencies
}

type Dependencies struct {
	// Pools
	IdentityPool shared.RepoPool[identity.Repo]
	AccessPool   shared.RepoPool[access.Repo]
	AuthPool     shared.RepoPool[auth.Repo]

	// Services
	IdentityProvider shared.IdentityProvider
	PwHasher         shared.PasswordHasher
	Mailer           shared.Mailer

	// Middleware
	*sharedhttp.Middleware
}

type Services struct {
}

func Raise(
	router chi.Router,
	deps Dependencies,
) {
	gtw := &gateway{
		deps,
	}

	router.Route("/user", gtw.userHandler)
	router.Route("/user", gtw.externalCredentialHandler)
}

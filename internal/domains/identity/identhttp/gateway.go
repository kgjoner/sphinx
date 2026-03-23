package identhttp

import (
	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

type gateway struct {
	Dependencies
}

type Dependencies struct {
	// Repository
	PGPool        shared.RepoPool
	IdentFactory  shared.RepoFactory[identity.Repo]
	AccessFactory shared.RepoFactory[access.Repo]
	AuthFactory   shared.RepoFactory[auth.Repo]

	// Services
	IdentityProvider shared.IdentityProvider
	PwHasher         shared.PasswordHasher
	Mailer           shared.Mailer

	// Middleware
	*sharedhttp.Middleware
}

func Raise(
	router chi.Router,
	deps Dependencies,
) {
	gtw := &gateway{
		deps,
	}

	router.Route("/user", func(r chi.Router) {
		gtw.userHandler(r)
		gtw.externalCredentialHandler(r)
	})
}

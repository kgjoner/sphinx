package authhttp

import (
	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/repositories/cache"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

type gateway struct {
	Dependencies
}

type Dependencies struct {
	// Repositories
	PGPool        shared.RepoPool
	CachePool     cache.Pool
	AuthFactory   shared.RepoFactory[auth.Repo]
	AccessFactory shared.RepoFactory[access.Repo]

	// Services
	IdentityProvider shared.IdentityProvider
	TokenProvider    auth.TokenProvider
	PwHasher         shared.PasswordHasher
	DataHasher       shared.DataHasher
	Challenger       auth.CodeChallenger
	Mailer           shared.Mailer

	// Middleware
	*sharedhttp.Middleware
}

func Raise(
	router chi.Router,
	deps Dependencies,
) {
	gtw := &gateway{
		Dependencies: deps,
	}

	router.Route("/auth", gtw.sessionHandlers)
	router.Route("/oauth", gtw.oauthHandlers)
}

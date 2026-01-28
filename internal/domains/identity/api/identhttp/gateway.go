package identhttp

import (
	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

type Gateway struct {
	Services
	BasePool shared.RepoPool[Repo]
	mid      *sharedhttp.Middleware
}

type Repo interface {
	identity.Repo
	access.Repo
	auth.Repo
}

type Services struct {
	IdentityProvider shared.IdentityProvider
	Hasher           shared.PasswordHasher
	Mailer           shared.Mailer
}

func Raise(
	router chi.Router,
	pools shared.RepoPool[Repo],
	services Services,
	mid *sharedhttp.Middleware,
) {
	gtw := &Gateway{
		BasePool: pools,
		Services: services,
		mid:      mid,
	}

	router.Route("/user", gtw.userHandler)
	router.Route("/user", gtw.externalCredentialHandler)
}

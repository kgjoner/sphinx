package authgtw

import (
	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/sphinx/internal/common"
)

type AuthGateway struct {
	common.Repos
	mid common.Middlewares
}

func Raise(router chi.Router, repos common.Repos) {
	authgtw := &AuthGateway{
		repos,
		common.Middlewares{
			AuthRepo: repos.AuthRepo,
		},
	}

	router.Mount("/account", authgtw.accountHandler())
}

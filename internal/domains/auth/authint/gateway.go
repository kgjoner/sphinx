// Auth Internal Client Gateway
package authint

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Gateway struct {
	Dependencies

	actor shared.Actor
	mu    sync.RWMutex
	keys  map[string]struct {
		publicKey any
		algorithm auth.Algorithm
	}
	lastFetch time.Time
}

type Dependencies struct {
	// Repositories
	PGPool      shared.RepoPool
	AuthFactory shared.RepoFactory[auth.Repo]

	// Services
	Encryptor      auth.Encryptor
	KeyProvisioner auth.KeyProvisioner
}

func Raise(deps Dependencies) *Gateway {
	return &Gateway{
		Dependencies: deps,
		actor: shared.Actor{
			ID:         uuid.Nil,
			Kind:       shared.KindSystem,
			AudienceID: uuid.MustParse(config.Env.ROOT_APP_ID),
			Permissions: []string{
				auth.PermKeysManage,
				auth.PermKeysReadAll,
			},
		},
		keys: make(map[string]struct {
			publicKey any
			algorithm auth.Algorithm
		}),
	}
}

package tokens

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type KeysGetter interface {
	// CurrentSigningKey returns the RS256 key that should be used for signing new tokens.
	CurrentSigningKey() (privateKey any, kid string, err error)
	// PublicKeyByKID retrieves a public key by its key ID for token validation.
	PublicKeyByKID(kid string) (publicKey any, algorithm auth.Algorithm, err error)
}

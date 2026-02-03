package authhttp

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/kgjoner/sphinx/internal/domains/auth"
)

// JWKSResponse represents the JSON Web Key Set response format.
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a single JSON Web Key.
type JWK struct {
	KeyType   string `json:"kty"` // Key type (RSA)
	Use       string `json:"use"` // Usage (sig for signature)
	Algorithm string `json:"alg"` // Algorithm (RS256)
	KeyID     string `json:"kid"` // Key ID
	N         string `json:"n"`   // RSA modulus (base64url)
	E         string `json:"e"`   // RSA exponent (base64url)
}

func jwksPresenter(keys []auth.SigningKeyPubView, keyProvisioner auth.KeyProvisioner, w http.ResponseWriter) {
	// Convert to JWKS format
	jwks := JWKSResponse{
		Keys: make([]JWK, 0, len(keys)),
	}

	for _, key := range keys {
		if key.Algorithm == auth.HS256 {
			continue
		}

		jwk, err := convertToJWK(key, keyProvisioner)
		if err != nil {
			// Log error but continue with other keys
			continue
		}
		jwks.Keys = append(jwks.Keys, jwk)
	}

	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jwks)
}

// convertToJWK converts a SigningKey to JWK format.
func convertToJWK(key auth.SigningKeyPubView, keyProvisioner auth.KeyProvisioner) (JWK, error) {
	// Parse public key from PEM
	publicKey, err := keyProvisioner.DecodePublic([]byte(key.PublicKey))
	if err != nil {
		return JWK{}, err
	}

	switch pubKeyTyped := publicKey.(type) {
	case *rsa.PublicKey:
		// Convert RSA modulus and exponent to base64url
		n := base64.RawURLEncoding.EncodeToString(pubKeyTyped.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKeyTyped.E)).Bytes())

		return JWK{
			KeyType:   "RSA",
			Use:       "sig",
			Algorithm: string(key.Algorithm),
			KeyID:     key.KeyID,
			N:         n,
			E:         e,
		}, nil
	default:
		return JWK{}, fmt.Errorf("unsupported public key type")
	}
}

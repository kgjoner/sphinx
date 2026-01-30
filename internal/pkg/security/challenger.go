package security

import (
	"crypto/sha256"
	"encoding/base64"
)

type CodeChallenger struct{}

func NewCodeChallenger() *CodeChallenger {
	return &CodeChallenger{}
}

/* ==========================================================================
  Challenger
========================================================================== */

func (c *CodeChallenger) DoesChallengeMatch(method string, challenge string, verifier string) bool {
	switch method {
	case "plain":
		return challenge == verifier
	case "S256":
		hash := sha256.Sum256([]byte(verifier))
		// Base64 URL encode (without padding)
		computed := base64.RawURLEncoding.EncodeToString(hash[:])
		return computed == challenge
	default:
		return false
	}
}

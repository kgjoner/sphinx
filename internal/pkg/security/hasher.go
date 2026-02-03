package security

import (
	"crypto/subtle"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

/* ==========================================================================
	PasswordHasher
========================================================================== */

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// HashPassword hashes a password using bcrypt.
func (h *BcryptHasher) HashPassword(str string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(str), 14)
	return string(hash)
}

func (h *BcryptHasher) DoesPasswordMatch(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

/* ==========================================================================
	DataHasher
========================================================================== */

type SHA3Hasher struct{}

func NewSHA3Hasher() *SHA3Hasher {
	return &SHA3Hasher{}
}

// HashData creates a SHA3-256 hash of the input string.
func (h *SHA3Hasher) HashData(data string) string {
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, []byte(data))
	return hex.EncodeToString(hash)
}

func (h *SHA3Hasher) DoesDataMatch(hashedData, data string) bool {
	inputHash := h.HashData(data)
	return subtle.ConstantTimeCompare([]byte(inputHash), []byte(hashedData)) == 1
}

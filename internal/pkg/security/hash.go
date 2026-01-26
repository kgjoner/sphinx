package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

type Hasher struct{}

func NewHasher() *Hasher {
	return &Hasher{}
}

// HashPassword hashes a password using bcrypt.
func (h *Hasher) HashPassword(str string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(str), 14)
	return string(hash)
}

func (h *Hasher) DoesPasswordMatch(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HashData creates a SHA3-256 hash of the input string.
func (h *Hasher) HashData(str string) string {
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, []byte(str))
	return fmt.Sprintf("%x", hash)
}

func (h *Hasher) DoesDataHashMatch(hashedData, data string) bool {
	return h.HashData(data) == hashedData
}

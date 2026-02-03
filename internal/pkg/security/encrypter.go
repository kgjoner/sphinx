package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrInvalidKeyLength  = errors.New("invalid key length")
)

/* ==========================================================================
	AESEncrypter - AES-256-GCM encryption for sensitive data at rest
========================================================================== */

type AESEncrypter struct {
	key []byte // 32 bytes for AES-256
}

// NewAESEncrypter creates a new AES-256-GCM encrypter.
// The masterKey is hashed to ensure it's exactly 32 bytes.
func NewAESEncrypter(masterKey string) (*AESEncrypter, error) {
	if masterKey == "" {
		return nil, ErrInvalidKeyLength
	}

	// Hash the master key to get exactly 32 bytes for AES-256
	hash := sha256.Sum256([]byte(masterKey))

	return &AESEncrypter{
		key: hash[:],
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext.
// The nonce is prepended to the ciphertext for decryption.
func (e *AESEncrypter) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	// GCM mode provides authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce (number used once) - must be unique for each encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for safe storage in database TEXT fields
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return []byte(encoded), nil
}

// Decrypt decrypts base64-encoded ciphertext encrypted with Encrypt.
func (e *AESEncrypter) Decrypt(ciphertext []byte) ([]byte, error) {
	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}

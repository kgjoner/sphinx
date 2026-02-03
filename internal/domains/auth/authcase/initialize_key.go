package authcase

import (
	"fmt"
	"time"

	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type InitializeKey struct {
	AuthRepo       auth.Repo
	KeyProvisioner auth.KeyProvisioner
	Encryptor      auth.Encryptor
}

type InitializeKeyInput struct {
	Actor shared.Actor
}

func (c InitializeKey) Execute(input InitializeKeyInput) (out bool, err error) {
	if err := auth.CanManageKeys(input.Actor); err != nil {
		return false, err
	}

	keys, err := c.AuthRepo.ListActiveSigningKeys()
	if err != nil {
		return false, err
	}

	// Return if have any RS256 keys
	for _, key := range keys {
		if key.Algorithm == auth.RS256 {
			return true, nil
		}
	}

	// Acquire distributed lock before generating
	release, err := c.AuthRepo.AcquireSigningKeyLock()
	if err != nil {
		// Another instance might be initializing, check again
		time.Sleep(100 * time.Millisecond)
		keys, err := c.AuthRepo.ListActiveSigningKeys()
		if err != nil {
			return false, err
		}

		for _, key := range keys {
			if key.Algorithm == auth.RS256 {
				// Key was created by another instance
				return true, nil
			}
		}

		return false, fmt.Errorf("failed to acquire lock for initialization: %w", err)
	}
	defer release()

	// Double-check after acquiring lock (another instance might have just created it)
	keys, err = c.AuthRepo.ListActiveSigningKeys()
	if err != nil {
		return false, err
	}
	for _, key := range keys {
		if key.Algorithm == auth.RS256 {
			// Key was created by another instance while we waited for lock
			return true, nil
		}
	}

	// Now safe to generate - we have the lock and confirmed no key exists
	_, err = c.generateKey(false)
	if err != nil {
		return false, fmt.Errorf("failed to initialize first RS256 key: %w", err)
	}

	return true, nil
}

// generateAndStoreKeyInternal does the actual key generation without locking.
// Used by functions that already hold the lock (InitializeIfNeeded, RotateKeys).
func (c InitializeKey) generateKey(shouldDelayActivation bool) (*auth.SigningKey, error) {
	privateKeyPEM, publicKeyPEM, err := c.KeyProvisioner.GeneratePair()
	if err != nil {
		return nil, err
	}

	// Encrypt private key
	encryptedPrivateKey, err := c.Encryptor.Encrypt(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	signingKey, err := auth.NewSigningKey(auth.SigningKeyCreationFields{
		Algorithm:             auth.RS256,
		PublicKey:             string(publicKeyPEM),
		PrivateKey:            string(encryptedPrivateKey),
		ShouldDelayActivation: shouldDelayActivation,
	})
	if err != nil {
		return nil, err
	}

	// Store in database
	err = c.AuthRepo.InsertSigningKey(signingKey)
	if err != nil {
		return nil, err
	}

	return signingKey, nil
}

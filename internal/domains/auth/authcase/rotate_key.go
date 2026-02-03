package authcase

import (
	"time"

	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type RotateKeys struct {
	AuthRepo       auth.Repo
	KeyProvisioner auth.KeyProvisioner
	Encryptor      auth.Encryptor
}

type RotateKeysInput struct {
	GracePeriod time.Duration
	Actor       shared.Actor
}

func (c RotateKeys) Execute(input RotateKeysInput) (out bool, err error) {
	if err := auth.CanManageKeys(input.Actor); err != nil {
		return false, err
	}

	release, err := c.AuthRepo.AcquireSigningKeyLock()
	if err != nil {
		// If another instance is rotating, that's fine - just skip this rotation
		return false, nil
	}
	defer release()

	// Mark expired keys as inactive
	err = c.AuthRepo.DeactivateExpiredKeys()
	if err != nil {
		return false, err
	}

	// Get current active RS256 keys
	activeKeys, err := c.AuthRepo.ListActiveSigningKeys()
	if err != nil {
		return false, err
	}

	// Generate new key
	i := InitializeKey{
		AuthRepo:       c.AuthRepo,
		KeyProvisioner: c.KeyProvisioner,
		Encryptor:      c.Encryptor,
	}
	newKey, err := i.generateKey(input.GracePeriod >= time.Hour)
	if err != nil {
		return false, err
	}

	for _, key := range activeKeys {
		if key.Algorithm == auth.RS256 && key.KeyID != newKey.KeyID {
			err = key.Rotate(input.GracePeriod)
			if err == nil {
				err = c.AuthRepo.UpdateSigningKey(*key)
				if err != nil {
					return false, err
				}
			}
		}
	}

	return true, nil
}

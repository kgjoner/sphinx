package authint

import (
	"context"
	"time"

	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
)

func (g *Gateway) InitializeKeysIfNeeded() error {
	repo := g.AuthFactory.NewDAO(
		context.Background(),
		g.PGPool.Connection(),
	)

	i := authcase.InitializeKey{
		AuthRepo:       repo,
		KeyProvisioner: g.KeyProvisioner,
		Encryptor:      g.Encryptor,
	}
	_, err := i.Execute(authcase.InitializeKeyInput{
		Actor: g.actor,
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) CurrentSigningKey() (privateKey any, kid string, err error) {
	repo := g.AuthFactory.NewDAO(
		context.Background(),
		g.PGPool.Connection(),
	)

	i := authcase.GetCurrentSigningKey{
		AuthRepo: repo,
	}
	currentKey, err := i.Execute(authcase.GetCurrentSigningKeyInput{
		Actor: g.actor,
	})
	if err != nil {
		return nil, "", err
	}

	decryptedKey, err := g.Encryptor.Decrypt([]byte(currentKey.PrivateKey))
	if err != nil {
		return nil, "", err
	}

	privateKey, err = g.KeyProvisioner.DecodePrivate(decryptedKey)
	if err != nil {
		return nil, "", err
	}

	return privateKey, currentKey.KeyID, nil
}

func (g *Gateway) PublicKeyByKID(kid string) (publicKey any, algorithm auth.Algorithm, err error) {
	g.mu.RLock()
	cached, exists := g.keys[kid]
	g.mu.RUnlock()
	if exists {
		return cached.publicKey, cached.algorithm, nil
	}

	i := authcase.GetKeyByKID{
		AuthRepo: g.AuthFactory.NewDAO(
			context.Background(),
			g.PGPool.Connection(),
		),
	}
	key, err := i.Execute(authcase.GetKeyByKIDInput{
		KID:   kid,
		Actor: g.actor,
	})
	if err != nil {
		return nil, "", err
	}

	publicKeyBytes := []byte(key.PublicKey)
	publicKey, err = g.KeyProvisioner.DecodePublic(publicKeyBytes)
	if err != nil {
		return nil, "", err
	}

	g.mu.Lock()
	g.keys[kid] = struct {
		publicKey any
		algorithm auth.Algorithm
	}{
		publicKey: publicKey,
		algorithm: key.Algorithm,
	}
	g.mu.Unlock()

	return publicKey, key.Algorithm, nil
}

func (g *Gateway) ShouldRotate() (bool, error) {
	repo := g.AuthFactory.NewDAO(
		context.Background(),
		g.PGPool.Connection(),
	)

	i := authcase.GetCurrentSigningKey{
		AuthRepo: repo,
	}
	currentKey, err := i.Execute(authcase.GetCurrentSigningKeyInput{
		Actor: g.actor,
	})
	if err != nil {
		return false, err
	}

	// Check if the newest key is older than the rotation period
	keyAge := time.Since(currentKey.CreatedAt)
	return keyAge >= time.Duration(config.Env.JWT.KEY_ROTATION_INTERVAL_HOURS)*time.Hour, nil
}

func (g *Gateway) RotateKeys() error {
	repo := g.AuthFactory.NewDAO(
		context.Background(),
		g.PGPool.Connection(),
	)

	i := authcase.RotateKeys{
		AuthRepo:       repo,
		KeyProvisioner: g.KeyProvisioner,
		Encryptor:      g.Encryptor,
	}
	_, err := i.Execute(authcase.RotateKeysInput{
		GracePeriod: time.Duration(config.Env.JWT.REFRESH_LIFETIME_IN_SEC) * time.Second,
		Actor:       g.actor,
	})
	if err != nil {
		return err
	}

	// Clear cached keys
	g.mu.Lock()
	g.keys = make(map[string]struct {
		publicKey any
		algorithm auth.Algorithm
	})
	g.mu.Unlock()

	return nil
}

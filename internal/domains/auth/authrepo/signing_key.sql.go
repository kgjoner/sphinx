package authrepo

import (
	"sync"
	"time"

	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type signingKeyCache struct {
	// In-memory cache of active keys for performance
	mu         sync.RWMutex
	activeKeys []*auth.SigningKey
	lastFetch  time.Time
	cacheTTL   time.Duration
}

func (c *signingKeyCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activeKeys = nil
	c.lastFetch = time.Time{}
}

var cache = signingKeyCache{
	activeKeys: nil,
	lastFetch:  time.Time{},
	cacheTTL:   1 * time.Hour,
}

func (q DAO) InsertSigningKey(key *auth.SigningKey) error {
	raw, exists := rawQueries["InsertSigningKey"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		key.ID,
		key.KeyID,
		key.Algorithm,
		key.PublicKey,
		key.PrivateKey,
		key.IsActive,
		key.ActivatesAt,
		key.ExpiresAt,
		key.RotatedAt,
	)

	if err == nil {
		cache.invalidate()
	}

	return err
}

func (q DAO) UpdateSigningKey(key auth.SigningKey) error {
	raw, exists := rawQueries["UpdateSigningKey"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw,
		key.ID,
		key.IsActive,
		key.ExpiresAt,
		key.RotatedAt,
	)

	if err == nil {
		cache.invalidate()
	}

	return err
}

func (q DAO) ListActiveSigningKeys() ([]*auth.SigningKey, error) {
	cache.mu.RLock()
	if time.Since(cache.lastFetch) < cache.cacheTTL && cache.activeKeys != nil {
		keys := cache.activeKeys
		cache.mu.RUnlock()
		return keys, nil
	}
	cache.mu.RUnlock()

	raw, exists := rawQueries["ListActiveSigningKeys"]
	if !exists {
		return nil, ErrNoQuery
	}

	rows, err := q.dbtx.QueryContext(q.ctx, raw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*auth.SigningKey
	for rows.Next() {
		var key auth.SigningKey
		err := rows.Scan(
			&key.ID,
			&key.KeyID,
			&key.Algorithm,
			&key.PublicKey,
			&key.PrivateKey,
			&key.IsActive,
			&key.CreatedAt,
			&key.ActivatesAt,
			&key.ExpiresAt,
			&key.RotatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	// Update cache
	cache.mu.Lock()
	cache.activeKeys = keys
	cache.lastFetch = time.Now()
	cache.mu.Unlock()

	return keys, rows.Err()
}

func (q DAO) GetSigningKeyByKID(kid string) (*auth.SigningKey, error) {
	raw, exists := rawQueries["GetSigningKeyByKID"]
	if !exists {
		return nil, ErrNoQuery
	}

	row := q.dbtx.QueryRowContext(q.ctx, raw, kid)

	var key auth.SigningKey
	err := row.Scan(
		&key.ID,
		&key.KeyID,
		&key.Algorithm,
		&key.PublicKey,
		&key.PrivateKey,
		&key.IsActive,
		&key.CreatedAt,
		&key.ActivatesAt,
		&key.ExpiresAt,
		&key.RotatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (q DAO) DeactivateExpiredKeys() error {
	raw, exists := rawQueries["DeactivateExpiredKeys"]
	if !exists {
		return ErrNoQuery
	}

	_, err := q.dbtx.ExecContext(q.ctx, raw)
	if err == nil {
		cache.invalidate()
	}

	return err
}

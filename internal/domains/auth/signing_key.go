package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/validator"
)

// SigningKey represents a cryptographic key used for JWT signing.
// For RS256, both PublicKey and PrivateKey contain PEM-encoded RSA keys.
// For HS256 (legacy), both fields are empty and the secret comes from config.
type SigningKey struct {
	ID          uuid.UUID `validate:"required"`
	KeyID       string    `validate:"required"` // kid - unique identifier used in JWT header
	Algorithm   Algorithm `validate:"required"`
	PrivateKey  string    // Encrypted PEM-encoded private key (empty for HS256)
	PublicKey   string    // PEM-encoded public key (empty for HS256)
	IsActive    bool
	CreatedAt   time.Time `validate:"required"`
	ActivatesAt time.Time `validate:"required"`
	ExpiresAt   htypes.NullTime
	RotatedAt   htypes.NullTime
}

/* ==============================================================================
	CONSTRUCTORS
============================================================================== */

type SigningKeyCreationFields struct {
	Algorithm             Algorithm
	PublicKey             string
	PrivateKey            string
	ShouldDelayActivation bool
	ExpiresAt             htypes.NullTime
}

func NewSigningKey(f SigningKeyCreationFields) (*SigningKey, error) {
	keyID := uuid.New()
	kid := keyID.String()

	now := time.Now()
	activationTime := now
	if f.ShouldDelayActivation {
		activationTime = now.Add(1 * time.Hour)
	}

	s := &SigningKey{
		ID:          keyID,
		KeyID:       kid,
		Algorithm:   f.Algorithm,
		PublicKey:   f.PublicKey,
		PrivateKey:  f.PrivateKey,
		IsActive:    true,
		CreatedAt:   now,
		ActivatesAt: activationTime,
		ExpiresAt:   f.ExpiresAt,
	}

	return s, validator.Validate(s)
}

/* ==============================================================================
	METHODS
============================================================================== */

func (s SigningKey) IsValid() error {
	errs := make(map[string]string)

	if s.Algorithm == RS256 {
		if s.PublicKey == "" {
			errs["public_key"] = "public_key is required for RS256 keys"
		}
		if s.PrivateKey == "" {
			errs["private_key"] = "private_key is required for RS256 keys"
		}
	}

	if len(errs) != 0 {
		return apperr.NewMapError(errs)
	}

	return nil
}

func (k *SigningKey) Rotate(gracePeriod time.Duration) error {
	if !k.RotatedAt.Time.IsZero() {
		return ErrRedundantRotation
	}

	now := time.Now()
	if !k.ExpiresAt.Time.IsZero() {
		if now.After(k.ExpiresAt.Time) {
			return ErrKeyAlreadyExpired
		}

		k.ExpiresAt = htypes.NullTime{Time: now.Add(gracePeriod)}
	}

	if gracePeriod == 0 {
		k.IsActive = false
	}

	k.RotatedAt = htypes.NullTime{Time: now}
	return nil
}

// IsExpired returns true if the key has expired.
func (k *SigningKey) IsExpired() bool {
	if k.ExpiresAt.Time.IsZero() {
		return false
	}
	return time.Now().After(k.ExpiresAt.Time)
}

// IsLegacyHS256 returns true if this is a legacy HS256 key entry.
func (k *SigningKey) IsLegacyHS256() bool {
	return k.Algorithm == HS256
}

// ShouldBeActive returns true if the key should still be active
// (not expired and currently marked as active).
func (k *SigningKey) ShouldBeActive() bool {
	return k.IsActive && !k.IsExpired()
}

// ShouldBeUsedForSigning returns true if the key should be used to sign new tokens.
// This checks both if it's active and if the activation time has passed.
func (k *SigningKey) ShouldBeUsedForSigning() bool {
	if !k.ShouldBeActive() {
		return false
	}

	// Check if activation time has passed
	return time.Now().After(k.ActivatesAt)
}

/* ==========================================================================
	VIEWS
========================================================================== */

type SigningKeyView struct {
	ID          uuid.UUID       `json:"id"`
	KeyID       string          `json:"keyId"`
	Algorithm   Algorithm       `json:"algorithm"`
	PrivateKey  string          `json:"-"`
	PublicKey   string          `json:"-"`
	IsActive    bool            `json:"isActive"`
	CreatedAt   time.Time       `json:"createdAt"`
	ActivatesAt time.Time       `json:"activatesAt"`
	ExpiresAt   htypes.NullTime `json:"expiresAt"`
	RotatedAt   htypes.NullTime `json:"rotatedAt"`
}

func (k SigningKey) View() SigningKeyView {
	return SigningKeyView{
		ID:          k.ID,
		KeyID:       k.KeyID,
		Algorithm:   k.Algorithm,
		PrivateKey:  k.PrivateKey,
		PublicKey:   k.PublicKey,
		IsActive:    k.IsActive,
		CreatedAt:   k.CreatedAt,
		ActivatesAt: k.ActivatesAt,
		ExpiresAt:   k.ExpiresAt,
		RotatedAt:   k.RotatedAt,
	}
}

type SigningKeyStatView struct {
	KeyID     string          `json:"keyId"`
	Algorithm Algorithm       `json:"algorithm"`
	PublicKey string          `json:"-"`
	IsActive  bool            `json:"isActive"`
	CreatedAt time.Time       `json:"createdAt"`
	ExpiresAt htypes.NullTime `json:"expiresAt"`
}

func (k SigningKey) StatView() SigningKeyStatView {
	return SigningKeyStatView{
		KeyID:     k.KeyID,
		Algorithm: k.Algorithm,
		PublicKey: k.PublicKey,
		IsActive:  k.IsActive,
		CreatedAt: k.CreatedAt,
		ExpiresAt: k.ExpiresAt,
	}
}

type SigningKeyPubView struct {
	KeyID     string    `json:"keyId"`
	Algorithm Algorithm `json:"algorithm"`
	PublicKey string    `json:"-"`
}

func (k SigningKey) PubView() SigningKeyPubView {
	return SigningKeyPubView{
		KeyID:     k.KeyID,
		Algorithm: k.Algorithm,
		PublicKey: k.PublicKey,
	}
}

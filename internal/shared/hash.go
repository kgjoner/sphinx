package shared

import (
	"database/sql/driver"
	"fmt"

	"github.com/kgjoner/cornucopia/v3/validator"
)

/* ==============================================================================
	Hashed Password
============================================================================== */

// HashedPassword represents a hashed password.
type HashedPassword struct {
	value string
}

func NewHashedPassword(plainPw string, hasher PasswordHasher) (*HashedPassword, error) {
	err := validator.Validate(plainPw, "required", "min=8", "max=128", "atLeastOne=letter number")
	if err != nil {
		return nil, err
	}

	hashed := hasher.HashPassword(plainPw)
	return &HashedPassword{value: hashed}, nil
}

func (h HashedPassword) IsZero() bool {
	return h.value == ""
}

func (h HashedPassword) String() string {
	return h.value
}

func (h HashedPassword) Value() (driver.Value, error) {
	return h.value, nil
}

func (h *HashedPassword) Scan(src interface{}) error {
	if src == nil {
		return ErrEmptyPassword
	}

	switch v := src.(type) {
	case string:
		h.value = v
	case []byte:
		h.value = string(v)
	default:
		return fmt.Errorf("unexpected type for HashedPassword: %T", src)
	}
	return nil
}

func (h *HashedPassword) UnmarshalText(text []byte) error {
	if text == nil {
		return ErrEmptyPassword
	}

	h.value = string(text)
	return nil
}

/* ==============================================================================
	Hashed Data
============================================================================== */

// HashedData represents a hashed data.
type HashedData struct {
	value string
}

func NewHashedData(plainData string, hasher DataHasher) (*HashedData, error) {
	if plainData == "" {
		return nil, ErrEmptyHashedData
	}

	hashed := hasher.HashData(plainData)
	return &HashedData{value: hashed}, nil
}

func (h HashedData) IsZero() bool {
	return h.value == ""
}

func (h HashedData) String() string {
	return h.value
}

func (h HashedData) Value() (driver.Value, error) {
	return h.value, nil
}

func (h *HashedData) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case string:
		h.value = v
	case []byte:
		h.value = string(v)
	default:
		return fmt.Errorf("unexpected type for HashedData: %T", src)
	}
	return nil
}

func (h *HashedData) UnmarshalText(text []byte) error {
	if text == nil {
		return nil
	}

	h.value = string(text)
	return nil
}

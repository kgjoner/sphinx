package shared

type AuthProof interface {
	ValidFor(target any) bool
}

/* ==============================================================================
	Password Proof
============================================================================== */

// PasswordProof is proof that a password matches the stored hash.
type PasswordProof struct {
	verified       bool
	matchedAgainst HashedPassword
}

// Check if the provided raw password matches the hashed password, and return a proof if it does.
func VerifyPassword(hashPw HashedPassword, rawPw string, hasher PasswordHasher) (*PasswordProof, error) {
	if match := hasher.DoesPasswordMatch(hashPw.String(), rawPw); !match {
		return nil, ErrInvalidCredentials
	}
	return &PasswordProof{verified: true, matchedAgainst: hashPw}, nil
}

// ValidFor checks if the proof is valid for the given HashedPassword.
func (p PasswordProof) ValidFor(hasPw any) bool {
	typed, ok := hasPw.(HashedPassword)
	if !ok {
		return false
	}

	return p.verified && p.matchedAgainst.String() == typed.String()
}

/* ==============================================================================
	Data Proof
============================================================================== */

// DataProof is proof that data matches the stored hash.
type DataProof struct {
	verified       bool
	matchedAgainst HashedData
}

// Check if the provided raw data matches the hashed data, and return a proof if it does.
func VerifyData(hashData HashedData, rawData string, hasher DataHasher) (*DataProof, error) {
	if match := hasher.DoesDataMatch(hashData.String(), rawData); !match {
		return nil, ErrInvalidCredentials
	}
	return &DataProof{verified: true, matchedAgainst: hashData}, nil
}

// ValidFor checks if the proof is valid for the given HashedData.
func (p DataProof) ValidFor(hasPw any) bool {
	typed, ok := hasPw.(HashedData)
	if !ok {
		return false
	}

	return p.verified && p.matchedAgainst.String() == typed.String()
}

/* ==============================================================================
	Code Proof
============================================================================== */

// CodeProof is proof that a code matches the stored one.
type CodeProof struct {
	verified       bool
	matchedAgainst string
}

// By now, it only stores the code in a valid proof. The real validation happens in ValidFor.
func VerifyCode(code string) (*CodeProof, error) {
	return &CodeProof{verified: true, matchedAgainst: code}, nil
}

// ValidFor checks if the proof is valid for the given code.
func (p CodeProof) ValidFor(hasPw any) bool {
	typed, ok := hasPw.(string)
	if !ok {
		return false
	}

	return p.verified && p.matchedAgainst == typed
}

package identity

/* ==============================================================================
	VerificationKind
============================================================================== */

type VerificationKind string

const (
	VerificationEmail         VerificationKind = "email"
	VerificationPhone         VerificationKind = "phone"
	VerificationPasswordReset VerificationKind = "password_reset"
)

func (s VerificationKind) Enumerate() any {
	return []VerificationKind{
		VerificationEmail,
		VerificationPhone,
		VerificationPasswordReset,
	}
}

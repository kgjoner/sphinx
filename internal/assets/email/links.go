package email

type Link struct {
	Key LinkKey
	URL string
}

type LinkKey string

const (
	ResetLink        LinkKey = "reset_link"
	CancelLink       LinkKey = "cancel_link"
	VerificationLink LinkKey = "verification_link"
)

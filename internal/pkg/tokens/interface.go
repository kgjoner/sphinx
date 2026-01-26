package tokens

type Provider interface {
	Generate(sub Subject) (*Tokens, error)
	Validate(token string) (*Subject, Intent, error)
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type Intent string

const (
	IntentAccess  Intent = "access"
	IntentRefresh Intent = "refresh"
)

func (i Intent) Enumerate() any {
	return []Intent{
		IntentAccess,
		IntentRefresh,
	}
}

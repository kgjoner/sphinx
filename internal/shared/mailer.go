package shared

import "github.com/kgjoner/cornucopia/v2/helpers/htypes"

type Mailer interface {
	Send(recipient htypes.Email, email Email, languages ...string) error
}

type Email interface {
	TemplateKey() string
}

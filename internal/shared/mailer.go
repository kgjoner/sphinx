package shared

import "github.com/kgjoner/cornucopia/v3/prim"

type Mailer interface {
	Send(recipient prim.Email, email Email, languages ...string) error
}

type Email interface {
	TemplateKey() string
}

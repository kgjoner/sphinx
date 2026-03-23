package shared

import (
	"fmt"
	"strings"

	"github.com/kgjoner/cornucopia/v3/apperr"
	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/cornucopia/v3/validator"
	"github.com/kgjoner/cornucopia/v3/sanitizer"
)

// Represents any entry of an user: email, phone, username or document.
type Entry string

func ParseEntry(str string) (Entry, error) {
	if str == "" {
		return "", nil
	}

	if strings.Contains(str, "@") {
		email, err := prim.ParseEmail(str)
		return Entry(email), err
	}

	if strings.HasPrefix(str, "+") {
		phone, err := prim.ParsePhoneNumber(str)
		return Entry(phone), err
	}

	// Try to parse a document even if it does not contain a colon.
	document, err := prim.ParseDocument(str)
	if err == nil || (strings.Contains(str, ":") || sanitizer.IsDigitOnly(str)) {
		// If it contains a colon or digit, it should be a document.
		return Entry(document), err
	}

	e := Entry(strings.ToLower(str))
	return e, e.IsValid()
}

func (e Entry) IsValid() error {
	var err error
	kind := e.Kind()
	switch kind {
	case "email":
		err = prim.Email(string(e)).IsValid()
	case "phone":
		err = prim.PhoneNumber(string(e)).IsValid()
	case "document":
		err = prim.Document(string(e)).IsValid()
	case "username":
		err = validator.Validate(string(e), "wordID", "atLeastOne=letter")
	}

	if err != nil {
		msg := fmt.Sprintf("invalid entry, identified like %v: %v", kind, err)
		return apperr.NewValidationError(msg)
	}

	return nil
}

func (e Entry) Kind() string {
	str := string(e)
	switch {
	case strings.Contains(str, "@"):
		return "email"
	case strings.Contains(str, "+"):
		return "phone"
	case strings.Contains(str, ":") || sanitizer.IsDigitOnly(str):
		return "document"
	default:
		return "username"
	}
}

func (e Entry) String() string {
	return string(e)
}

func (e *Entry) UnmarshalText(text []byte) error {
	parsed, err := ParseEntry(string(text))
	if err != nil {
		return err
	}
	*e = parsed
	return nil
}

package mailer

import (
	"html/template"

	"github.com/kgjoner/cornucopia/v3/prim"
)

type data struct {
	BaseData
	InvariantData
	VariantData
}

type VariantData struct {
	UserName        string
	NewEmail        string
	ValidationURL   string
	ResetURL        string
	CancelURL       string
	ConfirmationURL string
}

type InvariantData struct {
	AppName              string
	SupportEmail         string
	ClientBaseURL        string
	DataVerificationPath string
	PasswordResetPath    string

	AliasAddress prim.Email
	AliasName    string
}

type BaseData struct {
	PrimaryColor           string `json:"primaryColor"`
	PrimaryHoverColor      string `json:"primaryHoverColor"`
	ContentBackgroundColor string `json:"contentBackgroundColor"`
	BodyBackgroundColor    string `json:"bodyBackgroundColor"`
	LinkColor              string `json:"linkColor"`
	DividerColor           string `json:"dividerColor"`

	Header struct {
		Logo      string       `json:"logo"`
		Title     string       `json:"title"`
		Style     template.CSS `json:"style"`
		LogoStyle template.CSS `json:"logoStyle"`
	} `json:"header"`

	Footer struct {
		Text  string       `json:"text"`
		Style template.CSS `json:"style"`
	} `json:"footer"`
}

func (o *BaseData) PopulateZeros() {
	if o.PrimaryColor == "" {
		o.PrimaryColor = "#3498db"
	}

	if o.PrimaryHoverColor == "" {
		o.PrimaryHoverColor = "#34495e"
	}

	if o.LinkColor == "" {
		o.LinkColor = "#34495e"
	}

	if o.ContentBackgroundColor == "" {
		o.ContentBackgroundColor = "#ffffff"
	}

	if o.BodyBackgroundColor == "" {
		o.BodyBackgroundColor = "#f6f6f6"
	}

	if o.DividerColor == "" {
		o.DividerColor = "#e4e4e4"
	}
}

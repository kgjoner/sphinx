package email

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/kgjoner/cornucopia/v2/helpers/i18n"
	"github.com/kgjoner/hermes/pkg/hermes"
)

// Template returns the mail template for the given key and preferred languages (in order).
func Template(key TemplateKey, lngs ...string) emailTemplate {
	for _, lng := range lngs {
		res, exists := templates[i18n.Language(lng)]
		if exists {
			return res[key]
		}
	}

	return templates[i18n.Portuguese][key]
}

/* ================================================================================
	Keys
================================================================================ */

type TemplateKey string

const (
	Welcome                 TemplateKey = "welcome"
	PasswordReset           TemplateKey = "password_reset"
	PasswordUpdateNotice    TemplateKey = "password_update_notice"
	EmailUpdateNotice       TemplateKey = "email_update_notice"
	EmailUpdateConfirmation TemplateKey = "email_update_confirmation"
)

/* ================================================================================
	Handler
================================================================================ */

type emailTemplate struct {
	Subject struct {
		Content  string `json:"content"`
		template *template.Template
	} `json:"subject"`
	Body []struct {
		hermes.CustomTemplateDescriptor
		LinkKey  LinkKey `json:"link_key"`
		template *template.Template
	} `json:"body"`
}

// Parse parses the template strings into actual templates
func (a *emailTemplate) parse() {
	var err error
	a.Subject.template, err = template.New("subject").Parse(a.Subject.Content)
	if err != nil {
		panic(err)
	}

	for i, desc := range a.Body {
		desc.template, err = template.New(fmt.Sprintf("%v-%v", desc.Kind, i)).Parse(desc.Content)
		if err != nil {
			panic(err)
		}
		a.Body[i] = desc
	}
}

// Execute fills the template with the given params
func (a *emailTemplate) Execute(params Params) {
	a.Subject.Content = executeTemplate(a.Subject.template, params)

	for i, desc := range a.Body {
		desc.Content = executeTemplate(desc.template, params)
		a.Body[i] = desc
	}
}

// Descriptors fills the links in the body and returns the content descriptors
func (a emailTemplate) Descriptors(links ...Link) []hermes.CustomTemplateDescriptor {
	var res = []hermes.CustomTemplateDescriptor{}

	for _, desc := range a.Body {
		if desc.LinkKey != "" {
			for _, customLink := range links {
				if customLink.Key == desc.LinkKey {
					desc.Link = customLink.URL
				}
			}
		}
		res = append(res, desc.CustomTemplateDescriptor)
	}

	return res
}

func executeTemplate(templ *template.Template, params Params) string {
	var buf bytes.Buffer
	templ.Execute(&buf, params)
	return buf.String()
}

/* ================================================================================
	INIT
================================================================================ */

//go:embed *.json
var files embed.FS

type TemplateMap map[i18n.Language]map[TemplateKey]emailTemplate

var templates = TemplateMap{}

var acceptedLanguages = []i18n.Language{i18n.Portuguese}

func init() {
	for _, lng := range acceptedLanguages {
		rawFile, _ := files.ReadFile(string(lng) + ".json")
		var res map[TemplateKey]emailTemplate
		json.Unmarshal(rawFile, &res)

		for key, emailTemplate := range res {
			emailTemplate.parse()
			res[key] = emailTemplate
		}

		templates[lng] = res
	}
}

func MergeTemplates(extra TemplateMap) {
	for lng, tmplMap := range extra {
		if _, exists := templates[lng]; !exists {
			templates[lng] = map[TemplateKey]emailTemplate{}
		}

		for key, tmpl := range tmplMap {
			tmpl.parse()
			templates[lng][key] = tmpl
		}
	}
}

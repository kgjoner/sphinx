package i18n

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/kgjoner/cornucopia/helpers/i18n"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/config"
)

type resource struct {
	Mails map[string]resourceMail `json:"mails"`
}

type ResourceParams struct {
	AppName  string
	UserName string
}

type resourceMail struct {
	Subject struct {
		Content  string `json:"content"`
		Template *template.Template
	} `json:"subject"`
	Body []struct {
		hermes.CustomTemplateDescriptor
		Key      string `json:"key"`
		Template *template.Template
	} `json:"body"`
}

func (a *resourceMail) fillTemplate() {
	a.Subject.Template = parseTemplate("subject", a.Subject.Content)

	for i, desc := range a.Body {
		desc.Template = parseTemplate(fmt.Sprintf("%v-%v", desc.Kind, i), desc.Content)
		a.Body[i] = desc
	}
}

func (a *resourceMail) ParseContent(params ResourceParams) {
	if params.AppName == "" {
		params.AppName = config.Env.APP_NAME
	}

	a.Subject.Content = executeTemplate(a.Subject.Template, params)

	for i, desc := range a.Body {
		desc.Content = executeTemplate(desc.Template, params)
		a.Body[i] = desc
	}
}

type CustomLink struct {
	Key  string
	Link string
}

func (a resourceMail) FormatBody(customLinks ...CustomLink) []hermes.CustomTemplateDescriptor {
	var res = []hermes.CustomTemplateDescriptor{}

	for _, desc := range a.Body {
		if desc.Key != "" {
			for _, customLink := range customLinks {
				if customLink.Key == desc.Key {
					desc.Link = customLink.Link
				}
			}
		}
		res = append(res, desc.CustomTemplateDescriptor)
	}

	return res
}

func parseTemplate(name string, target string) *template.Template {
	t, err := template.New(name).Parse(target)
	if err != nil {
		panic(err)
	}
	return t
}

func executeTemplate(templ *template.Template, params ResourceParams) string {
	var buf bytes.Buffer
	templ.Execute(&buf, params)
	return buf.String()
}

/*
================================================================================

	INIT

================================================================================
*/
type resourceMap map[i18n.Language]resource

//go:embed *.json
var files embed.FS

var AcceptedLanguages = []i18n.Language{i18n.LanguageValues.PT_BR, i18n.LanguageValues.EN_US}
var Resources = resourceMap{}

func init() {
	for _, lng := range AcceptedLanguages {
		rawFile, _ := files.ReadFile(string(lng) + ".json")
		var res resource
		json.Unmarshal(rawFile, &res)

		for _, resourceMail := range res.Mails {
			resourceMail.fillTemplate()
		}

		Resources[lng] = res
	}
}

func Resource(lngs []string) resource {
	for _, lng := range lngs {
		res, exists := Resources[i18n.Language(lng)]
		if exists {
			return res
		}
	}

	return Resources[i18n.Language(config.Env.FALLBACK_LANGUAGE)]
}

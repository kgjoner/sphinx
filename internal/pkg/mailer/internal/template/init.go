package templates

import (
	"embed"
)

type Templates map[string]Registry

//go:embed base/*.html
var baseHTML embed.FS

//go:embed en/*.html
var enHTML embed.FS

//go:embed pt/*.html
var ptHTML embed.FS

var languages = map[string]embed.FS{
	"en": enHTML,
	"pt": ptHTML,
}

var keys = []Key{
	"welcome",
	"password_reset",
	"password_update_notice",
	"email_update_confirmation",
	"email_update_notice",
}

func New(overrides map[string]map[Key]string) (Templates, error) {
	templates := make(Templates)

	// 1. Load the Base template
	baseRaw, err := baseHTML.ReadFile("base/template.html")
	if err != nil {
		return nil, err
	}

	enSignatureRaw, err := baseHTML.ReadFile("base/en_signature.html")
	if err != nil {
		return nil, err
	}

	ptSignatureRaw, err := baseHTML.ReadFile("base/pt_signature.html")
	if err != nil {
		return nil, err
	}

	signatures := map[string]string{
		"en": string(enSignatureRaw),
		"pt": string(ptSignatureRaw),
	}

	// 2. Load the specific email kinds
	for lng, htmlFS := range languages {
		emailKinds := make(map[Key]string)
		for _, key := range keys {
			if overrides[lng] != nil {
				if content, ok := overrides[lng][key]; ok {
					emailKinds[key] = content
					continue
				}
			}

			content, err := htmlFS.ReadFile(lng + "/" + string(key) + ".html")
			if err != nil {
				return nil, err
			}
			emailKinds[key] = string(content)
		}

		// 3. Create the registry for this language
		registry, err := newRegistry(string(baseRaw), signatures[lng], emailKinds)
		if err != nil {
			return nil, err
		}

		templates[lng] = *registry
	}

	return templates, nil
}
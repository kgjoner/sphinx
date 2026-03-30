package templates

import (
	"bytes"
	"fmt"
	"html/template"
)

var ErrTemplateNotFound = fmt.Errorf("template not found on registry")

type Key string

type Registry struct {
	cache map[Key]*template.Template
}

func newRegistry(baseRaw string, signatureRaw string, emailKinds map[Key]string) (*Registry, error) {
	registry := &Registry{
		cache: make(map[Key]*template.Template),
	}

	// 1. Parse the Base once as a "Master"
	master, err := template.New("base").Parse(baseRaw)
	if err != nil {
		return nil, err
	}

	for name, content := range emailKinds {
		// 2. Clone the master so we don't mess with the original Base
		t, err := master.Clone()
		if err != nil {
			return nil, err
		}
		
		// 3. Parse the signature template into the clone
		t, err = t.Parse(signatureRaw)
		if err != nil {
			return nil, err
		}
		
		// 4. Parse the specific email kind into the clone
		t, err = t.Parse(content)
		if err != nil {
			return nil, err
		}

		registry.cache[name] = t
	}

	return registry, nil
}

func (r *Registry) Execute(key Key, data any) (subject string, body string, err error) {
	tmpl, ok := r.cache[key]
	if !ok {
		return "", "", ErrTemplateNotFound
	}

	var subBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&subBuf, "subject", data); err != nil {
		return "", "", err
	}

	var bodyBuf bytes.Buffer
	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", err
	}

	return subBuf.String(), bodyBuf.String(), nil
}

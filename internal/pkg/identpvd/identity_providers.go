package identpvd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kgjoner/cornucopia/v3/apperr"
	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/cornucopia/v3/validator"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Providers struct {
	providers map[string]Provider
}

func NewProviders(raw []byte) (*Providers, error) {
	if len(raw) == 0 {
		return &Providers{
			providers: make(map[string]Provider),
		}, nil
	}

	var providers []Provider
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&providers); err != nil {
		return nil, fmt.Errorf("failed to parse external identity providers: %w", err)
	}

	providersMap := make(map[string]Provider)
	for _, provider := range providers {
		err := validator.Validate(provider)
		if err != nil {
			return nil, fmt.Errorf("invalid external identity provider '%s': %w", provider.Name, err)
		}
		providersMap[provider.Name] = provider
	}

	return &Providers{
		providers: providersMap,
	}, nil
}

type Provider struct {
	Name           string            `validate:"required"`
	URL            string            `validate:"required"`
	Method         string            `json:",omitempty"` // Default is GET
	DefaultHeaders map[string]string `json:",omitempty"`
	DefaultParams  map[string]string `json:",omitempty"`
	DefaultBody    map[string]string `json:",omitempty"`

	// JSON path to extract subject ID from response
	SubjectIDPath string `json:",omitempty" validate:"required"`
	EmailPath     string `json:",omitempty"`
	AliasPath     string `json:",omitempty"`
}

func (p *Providers) Authenticate(input shared.IdentityProviderInput) (subject *shared.ExternalSubject, err error) {
	e, ok := p.providers[input.ProviderName]
	if !ok {
		return nil, shared.ErrInvalidExternalSubject
	}

	if e.Method == "" {
		e.Method = http.MethodGet
	}

	// Prepare body if needed
	var body []byte
	if (e.Method == http.MethodPost || e.Method == http.MethodPut) && (len(e.DefaultBody) > 0 || len(input.Body) > 0) {
		bodyMap := make(map[string]string)
		for k, v := range e.DefaultBody {
			bodyMap[k] = v
		}
		for k, v := range input.Body {
			bodyMap[k] = v
		}
		body, err = json.Marshal(bodyMap)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(e.Method, e.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range e.DefaultHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", input.Authorization)

	// Set query parameters
	q := req.URL.Query()
	for k, v := range e.DefaultParams {
		q.Set(k, v)
	}
	for k, v := range input.Params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperr.NewUnauthorizedError("authentication failed")
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract subject ID using the provided path
	subjectID, err := extractValueFromPath(responseBody, e.SubjectIDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract subject ID: %w", err)
	}

	var email prim.Email
	if e.EmailPath != "" {
		strEmail, err := extractValueFromPath(responseBody, e.EmailPath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract email: %w", err)
		}

		email, err = prim.ParseEmail(strEmail)
		if err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	var alias string
	if e.AliasPath != "" {
		alias, err = extractValueFromPath(responseBody, e.AliasPath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract alias: %w", err)
		}
	} else if email != "" {
		alias = email.String()
	}

	return &shared.ExternalSubject{
		ProviderName: e.Name,
		ID:           subjectID,
		Kind:         shared.KindUser,
		Email:        email,
		Alias:        alias,
	}, nil
}

func extractValueFromPath(data map[string]interface{}, path string) (string, error) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			value, ok := current[part]
			if !ok {
				return "", fmt.Errorf("key '%s' not found in response at path: %s", part, path)
			}
			if value == nil {
				return "", fmt.Errorf("value is nil at path: %s", path)
			}

			ans, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("expected string at '%s' but got %T at path: %s", part, value, path)
			}

			return ans, nil
		} else {
			// Intermediate part - must be another object
			value, ok := current[part]
			if !ok {
				return "", fmt.Errorf("key '%s' not found in response at path: %s", part, path)
			}

			nextMap, ok := value.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("expected object at '%s' but got %T at path: %s", part, value, path)
			}
			current = nextMap
		}
	}

	return "", fmt.Errorf("empty path provided")
}

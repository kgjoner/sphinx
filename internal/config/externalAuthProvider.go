package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
)

type ExternalAuthProvider struct {
	Name           string `validate:"required"`
	URL            string `validate:"required"`
	Method         string // Default is GET
	DefaultHeaders map[string]string
	DefaultParams  map[string]string
	DefaultBody    map[string]string

	// JSON path to extract subject ID from response
	SubjectIDPath string `validate:"required"`
	EmailPath     string
}

type ExternalAuthInput struct {
	AuthorizationHeader string
	Params              map[string]string
	Body                map[string]string
}

func (e ExternalAuthProvider) Authenticate(input ExternalAuthInput) (subject *ExternalSubject, err error) {
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
	req.Header.Set("Authorization", input.AuthorizationHeader)

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
		return nil, apperr.NewUnauthorizedError("authentication failed: %s")
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract subject ID using the provided path
	subjectID, err := e.extractValueFromPath(responseBody, e.SubjectIDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract subject ID: %w", err)
	}

	var email htypes.Email
	if e.EmailPath != "" {
		strEmail, err := e.extractValueFromPath(responseBody, e.EmailPath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract email: %w", err)
		}

		email, err = htypes.ParseEmail(strEmail)
		if err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	return &ExternalSubject{
		ProviderName: e.Name,
		ID:           subjectID,
		Email:        email,
		isAuthed:     true,
	}, nil
}

func (e ExternalAuthProvider) extractValueFromPath(data map[string]interface{}, path string) (string, error) {
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

type ExternalSubject struct {
	ProviderName string
	ID           string
	Email        htypes.Email

	isAuthed bool
}

func (s *ExternalSubject) IsAuthenticated() bool {
	return s.isAuthed
}

package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/vrischmann/envconfig"
)

var Env struct {
	// If "development" is set, SMTP connection will be insecure (TLS disabled) and CORS will allow all origins
	// (it may be overridden by ALLOWED_ORIGINS env variable). Use only for local development.
	APP_ENV      string `envconfig:"default=production"`
	SCHEME       string `envconfig:"default=http"`
	HOST         string `envconfig:"default=localhost:8080"`
	APP_VERSION  string `envconfig:"default=v1.0.0"`
	// BASE_PATH is the base path for all API endpoints, it will be automatically suffixed with the major 
	// version from APP_VERSION.
	BASE_PATH    string `envconfig:"optional"`
	DATABASE_URL string
	REDIS_URL    string
	ROOT_APP_ID  string `envconfig:"default=80cadd74-5ccd-41c4-9938-3c8961be04db"`
	// Comma-separated list of allowed origins for CORS. If not set, it will default to the CLIENT.BASE_URL.
	// In development environment, it will default to allowing all origins.
	ALLOWED_ORIGINS []string `envconfig:"optional"`

	// Set 0 for disabling concurrent sessions control
	MAX_CONCURRENT_SESSIONS    int `envconfig:"default=0"`
	AUTH_GRANT_LIFETIME_IN_SEC int `envconfig:"default=300"`
	JWT                        struct {
		ACCESS_LIFETIME_IN_SEC  int    `envconfig:"default=900"`
		REFRESH_LIFETIME_IN_SEC int    `envconfig:"default=172800"`
		ALGORITHM               string `envconfig:"default=RS256"` // RS256 or HS256 (legacy)
		SECRET                  string // Legacy HS256 secret (for initial grace period)
		ENCRYPTION_KEY          string
		// How often to rotate keys (in hours), default 1 year. Set 0 to disable automatic rotation.
		KEY_ROTATION_INTERVAL_HOURS int `envconfig:"default=8760"`
	}

	CLIENT struct {
		BASE_URL          string `envconfig:"default=http://localhost:3000"`
		DATA_VERIFICATION string `envconfig:"default=/verification"`
		PASSWORD_RESET    string `envconfig:"default=/password/reset"`
	}

	APP_NAME          string `envconfig:"default=Sphinx"`
	APP_STYLE_URL     string `envconfig:"optional"` // if none is provided, the default style will be used (see assets/style/style.go)
	APP_LOGO_URL      string `envconfig:"optional"` // if none is provided, the default logo will be used (see assets/img/logo.svg)
	SUPPORT_EMAIL     string `envconfig:"default=support@example.test"`
	FALLBACK_LANGUAGE string `envconfig:"default=pt"`

	SMTP struct {
		USERNAME string
		PASSWORD string
		HOST     string
		PORT     string
	}

	// Used to protect the Swagger UI with basic authentication. The expected format is a JSON
	// object where keys are usernames and values are their respective passwords. For example:
	// {"admin": "password123", "user1": "passw0rd"}.
	//
	// If not provided, the Swagger UI will be accessible without authentication.
	SWAGGER_AUTH map[string]string `envconfig:"-" json:",omitempty"`
	// Overwrite default email templates with custom ones. The expected format is a JSON object
	// where first level keys are the language codes (e.g., "en", "pt"), the second level keys
	// are the template keys (e.g., "welcome", "password_reset"), and the values are the template
	// content as string.
	//
	// Check internal/pkg/mailer/internal/assets for reference on the expected structure.
	EMAIL_TEMPLATES []byte `envconfig:"-" json:",omitempty"`
	// Configure external authentication providers (e.g., Google, Facebook) with their respective
	// validation endpoints. The expected format is a JSON array of objects containing the necessary
	// configuration for each provider. It is optional, so if not provided, no external auth will
	// be possible. Each provider must have a unique name.
	//
	// Check internal/server/identpvd/identity_providers.go for reference on the expected structure.
	//
	// Use with caution, only with trusted providers, as this may open security vulnerabilities.
	EXTERNAL_AUTH_PROVIDERS []byte `envconfig:"-" json:",omitempty"`
}

func Must() {
	if err := envconfig.Init(&Env); err != nil {
		panic(err)
	}

	// Handle SWAGGER_AUTH JSON parsing manually
	if swaggerAuthJSON := os.Getenv("SWAGGER_AUTH"); swaggerAuthJSON != "" {
		if err := json.Unmarshal([]byte(swaggerAuthJSON), &Env.SWAGGER_AUTH); err != nil {
			panic("failed to parse SWAGGER_AUTH JSON: " + err.Error())
		}
	}

	// Handle EMAIL_TEMPLATES JSON parsing manually
	emailTemplates := []byte(os.Getenv("EMAIL_TEMPLATES"))
	Env.EMAIL_TEMPLATES = emailTemplates

	// Handle EXTERNAL_AUTH_PROVIDERS JSON parsing manually
	externalAuthProviders := []byte(os.Getenv("EXTERNAL_AUTH_PROVIDERS"))
	Env.EXTERNAL_AUTH_PROVIDERS = externalAuthProviders

	if !strings.HasPrefix(Env.APP_VERSION, "v") {
		Env.APP_VERSION = "v" + Env.APP_VERSION
	}

	baseVersion := strings.Split(Env.APP_VERSION, "-")[0]
	semver := strings.Split(baseVersion, ".")
	if len(semver) != 3 {
		panic("APP_VERSION env must be in form of semantic versioning")
	}

	Env.BASE_PATH += "/" + semver[0]
}

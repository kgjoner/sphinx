package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/vrischmann/envconfig"
)

var Env struct {
	SCHEME       string `envconfig:"default=http"`
	HOST         string `envconfig:"default=localhost:8080"`
	APP_VERSION  string `envconfig:"default=v1.0.0"`
	DATABASE_URL string `envconfig:"default=postgres://postgres:postgres@localhost:5432/sphinx?sslmode=disable"`
	REDIS_URL    string `envconfig:"default=redis://localhost:6379/0"`
	ROOT_APP_ID  string `envconfig:"default=80cadd74-5ccd-41c4-9938-3c8961be04db"`

	//Set 0 for disabling concurrent sessions control
	MAX_CONCURRENT_SESSIONS    int `envconfig:"default=0"`
	AUTH_GRANT_LIFETIME_IN_SEC int `envconfig:"default=300"`
	JWT                        struct {
		SECRET                  string `envconfig:"default=topsecret"`
		ACCESS_LIFETIME_IN_SEC  int    `envconfig:"default=900"`
		REFRESH_LIFETIME_IN_SEC int    `envconfig:"default=172800"`
	}

	CLIENT struct {
		BASE_URL          string `envconfig:"default=http://localhost:3000"`
		DATA_VERIFICATION string `envconfig:"default=/verification"`
		PASSWORD_RESET    string `envconfig:"default=/password/reset"`
	}

	APP_NAME          string `envconfig:"default=Sphinx"`
	APP_STYLE_URL     string `envconfig:"optional"` // if none is provided, the default style will be used (see assets/style/style.go)
	APP_LOGO_URL      string `envconfig:"optional"` // if none is provided, the default logo will be used (see assets/img/logo.svg)
	SUPPORT_EMAIL     string `envconfig:"default=support@example.com"`
	FALLBACK_LANGUAGE string `envconfig:"default=pt-br"`
	HERMES            struct {
		BASE_URL string `envconfig:"default=http://localhost:8081/v1"`
		API_KEY  string `envconfig:"default=topsecret"`
	}

	SWAGGER_AUTH    map[string]string `envconfig:"-" json:",omitempty"`
	EMAIL_TEMPLATES []byte            `envconfig:"-" json:",omitempty"`
	// Used for integrating with third-party identity providers.
	// It is OPTIONAL, so if not provided, no external auth will be possible.
	// Each provider must have a unique name.
	//
	// Use with caution, only with trusted providers, as this may open security vulnerabilities.
	//
	// See documentation for more details.
	EXTERNAL_AUTH_PROVIDERS []byte `envconfig:"-" json:",omitempty"`
}

var BASE_PATH string

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

	BASE_PATH = "/" + semver[0]
}

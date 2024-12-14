package config

import (
	"github.com/vrischmann/envconfig"
)

var Env struct {
	HOST           string `envconfig:"default=localhost:8080"`
	DATABASE_URL   string `envconfig:"default=postgres://postgres:postgres@db:5432/sphynx?sslmode=disable&pool_max_conns=20"`
	ROOT_APP_TOKEN string `envconfig:"default=80cadd74-5ccd-41c4-9938-3c8961be04db"`

	//Set 0 for disabling concurrent sessions control
	MAX_CONCURRENT_SESSIONS int `envconfig:"default=0"`
	OAUTH_LIFETIME_IN_SEC   int `envconfig:"default=300"`
	JWT                     struct {
		SECRET                  string `envconfig:"default=topsecret"`
		ACCESS_LIFETIME_IN_SEC  int    `envconfig:"default=300"`
		REFRESH_LIFETIME_IN_SEC int    `envconfig:"default=172800"`
	}

	CLIENT_URI struct {
		DATA_VERIFICATION string `envconfig:"default=localhost:8080/verification"`
		PASSWORD_RESET    string `envconfig:"default=localhost:8080/password/reset"`
	}

	APP_NAME          string `envconfig:"default=Sphinx"`
	SUPPORT_EMAIL      string `envconfig:"default=support@example.com"`
	FALLBACK_LANGUAGE string `envconfig:"default=pt-br"`
	HERMES struct {
		BASE_URL string `envconfig:"default=https://hermes.example.com"`
		API_KEY  string `envconfig:"default=topsecret"`
	}
}

func Must() {
	if err := envconfig.Init(&Env); err != nil {
		panic(err)
	}
}

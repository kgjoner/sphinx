package config

import (
	"github.com/vrischmann/envconfig"
)

var Environment struct {
	APP_NAME       string `envconfig:"default=Sphinx"`
	DATABASE_URL   string `envconfig:"default=postgres://postgres:postgres@db:5432/sphynx?sslmode=disable&pool_max_conns=20"`
	ROOT_APP_TOKEN string `envconfig:"default=80cadd74-5ccd-41c4-9938-3c8961be04db"`
	SWAGGER_HOST   string `envconfig:"default=localhost:8080"`
	//Set 0 for disabling concurrent sessions control
	MAX_CONCURRENT_SESSIONS int    `envconfig:"default=0"`
	FALLBACK_LANGUAGE       string `envconfig:"default=pt-br"`
	JWT                     struct {
		SECRET                   string `envconfig:"default=topsecret"`
		ACCESS_LIFE_TIME_IN_SEC  int    `envconfig:"default=300"`
		REFRESH_LIFE_TIME_IN_SEC int    `envconfig:"default=172800"`
	}
	CLIENT_URI struct {
		DATA_VERIFICATION string `envconfig:"default=localhost:8080/verification"`
		PASSWORD_RESET    string `envconfig:"default=localhost:8080/password-reset"`
	}
	HERMES struct {
		BASE_URL string
		API_KEY  string `envconfig:"default=topsecret"`
	}
}

func Must() {
	if err := envconfig.Init(&Environment); err != nil {
		panic(err)
	}
}

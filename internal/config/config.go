package config

import (
	"github.com/vrischmann/envconfig"
)

var Env struct {
	HOST         string `envconfig:"default=localhost:8080"`
	DATABASE_URL string `envconfig:"default=postgres://postgres:postgres@db:5432/sphynx?sslmode=disable&pool_max_conns=20"`
	REDIS_URL    string `envconfig:"default=redis://redis:redis@rdb:6379/0"`
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
		BASE_URL          string `envconfig:"default=localhost:8080"`
		DATA_VERIFICATION string `envconfig:"default=/verification"`
		PASSWORD_RESET    string `envconfig:"default=/password/reset"`
	}

	APP_NAME          string `envconfig:"default=Sphinx"`
	APP_STYLE_URL     string // if none is provided, the default style will be used (see assets/style/style.go)
	APP_LOGO_URL      string // if none is provided, the default logo will be used (see assets/img/logo.svg)
	SUPPORT_EMAIL     string `envconfig:"default=support@example.com"`
	FALLBACK_LANGUAGE string `envconfig:"default=pt-br"`
	HERMES            struct {
		BASE_URL string `envconfig:"default=https://hermes.example.com"`
		API_KEY  string `envconfig:"default=topsecret"`
	}
}

func Must() {
	if err := envconfig.Init(&Env); err != nil {
		panic(err)
	}
}

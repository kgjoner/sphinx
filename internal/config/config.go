package config

import (
	"github.com/vrischmann/envconfig"
)

var Environment struct {
	DATABASE_URL string `envconfig:"default=postgres://postgres:postgres@db:5432/sphynx?sslmode=disable&pool_max_conns=20"`
	SWAGGER_HOST string `envconfig:"default=localhost:8080"`
	//Set 0 for disabling concurrent sessions control
	MAX_CONCURRENT_SESSIONS int `envconfig:"default=0"`
	JWT                     struct {
		SECRET                   string `envconfig:"default=topsecret"`
		ACCESS_LIFE_TIME_IN_SEC  int    `envconfig:"default=300"`
		REFRESH_LIFE_TIME_IN_SEC int    `envconfig:"default=172800"`
	}
}

func Must() {
	if err := envconfig.Init(&Environment); err != nil {
		panic(err)
	}
}

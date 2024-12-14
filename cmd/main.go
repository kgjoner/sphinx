package main

import (
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/server"
)

func main() {
	config.Must()
	server.New().Setup().Start()
}

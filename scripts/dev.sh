#! /bin/bash

DOCKER_BUILDKIT=1 docker build -f Dockerfile . -t kgjoner/sphinx:dev --ssh default=$SSH_AUTH_SOCK
docker-compose up
migrate -path postgres/migrations -database postgres://postgres:postgres@localhost:5432/sphinx?sslmode=disable up
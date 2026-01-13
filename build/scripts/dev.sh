#! /bin/bash

set -e
eval "$(ssh-agent)"
ssh-add ~/.ssh/id_ed25519
docker build -f Dockerfile . -t kgjoner/sphinx:dev --ssh default=$SSH_AUTH_SOCK
docker-compose up
migrate -path internal/repositories/base/migrations -database postgres://postgres:postgres@localhost:5432/sphinx?sslmode=disable up
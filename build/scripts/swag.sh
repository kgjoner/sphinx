#! /bin/bash

swag fmt
swag init -g server.go --dir internal/server,internal/domains/auth/gateway --parseDependency --parseInternal
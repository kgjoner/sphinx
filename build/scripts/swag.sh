#! /bin/bash

set -e

swag fmt
DIRS=internal/server,$(find internal/domains -type d -path "*/*http" | paste -sd ",")
echo "Generating swag docs for dirs: $DIRS"
swag init -g server.go --dir $DIRS --parseDependency --parseInternal
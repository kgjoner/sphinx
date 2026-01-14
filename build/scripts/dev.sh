#! /bin/bash

set -e

docker-compose up -d

echo "Waiting for database to be ready..."
until docker exec sphinx-pg pg_isready -U postgres > /dev/null 2>&1; do
  echo "Database not ready yet, waiting..."
  sleep 1
done;
echo "Database is ready!"

set -a
source .env
set +a

go run cmd/migrate/main.go
go run cmd/sphinx/main.go
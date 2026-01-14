#! /bin/bash

set -e

docker-compose up -d

echo "Waiting for database to be ready..."
until docker exec sphinx-pg pg_isready -U postgres > /dev/null 2>&1; do
  echo "Database not ready yet, waiting..."
  sleep 1
done;
echo "Database is ready!"

if [ -f .env ]; then
  set -a
  source .env
  set +a
else
  echo ".env not found — continuing with default environment variables"
fi

go run cmd/migrate/main.go
go run cmd/sphinx/main.go
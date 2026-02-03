#! /bin/bash

set -e

echo "🚀 Setting up dev environment..."

if [ -f .env ]; then
  set -a
  source .env
  set +a
else
  echo ".env not found — continuing with default environment variables"
fi

# Start docker-compose services
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
  if docker exec sphinx-pg pg_isready -U postgres > /dev/null 2>&1; then
    echo "✅ PostgreSQL is ready!"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ PostgreSQL failed to start in time"
    docker-compose -f docker-compose.yaml logs pg
    exit 1
  fi
  sleep 1
done

go run cmd/migrate/main.go

echo "✅ All auxiliary services are ready!"
echo ""
echo "📝 Dev environment details:"
echo "   PostgreSQL: localhost:5432"
echo "   Redis: localhost:6379"
echo "   Mailhog UI: http://localhost:8025"
echo "   Hermes: http://localhost:8081"
echo ""
echo "🛑 To stop services:"
echo "   make dev-down"

go run cmd/sphinx/main.go
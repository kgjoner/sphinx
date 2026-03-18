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

if [ -n "$DOCKER_REGISTRY" ]; then
  echo -e "🐳 Using Docker registry: $DOCKER_REGISTRY\n"
  if [[ "$DOCKER_REGISTRY" != */ ]]; then
    DOCKER_REGISTRY="$DOCKER_REGISTRY/"
  fi
else 
  echo -e "🐳 Using DockerHub as default registry\n"
fi

# Start docker compose services
docker compose up -d

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
  if docker exec sphinx-pg pg_isready -U postgres > /dev/null 2>&1; then
    echo -e "✅ PostgreSQL is ready!\n"
    break
  fi
  if [ $i -eq 30 ]; then
    echo -e "❌ PostgreSQL failed to start in time\n"
    docker compose -f docker-compose.yaml logs pg
    exit 1
  fi
  sleep 1
done

go run cmd/migrate/main.go

HAS_DATA=$(psql $DATABASE_URL -tAc "SELECT 1 FROM link LIMIT 1;")
if [ "$HAS_DATA" != "1" ]; then
  go run cmd/testseed/main.go clean
  go run cmd/testseed/main.go seed  
else
  echo "Sphinx data already exists, skipping seeding."
fi

# Extract major version from $APP_VERSION
MAJOR_VERSION=$(echo ${APP_VERSION:-v1.0.0} | cut -d. -f1)

echo ""
echo "✅ All auxiliary services are ready!"
echo ""
echo "📝 Dev environment details:"
echo "   Entrypoint: http://localhost:8080/${MAJOR_VERSION}"
echo "   Swagger: http://localhost:8080/${MAJOR_VERSION}/docs/swagger.json"
echo "   PostgreSQL: localhost:${DB_PORT:-5432}"
echo "   Redis: localhost:${REDIS_PORT:-6379}"
echo "   Mailhog UI: http://localhost:${MAILHOG_WEB_PORT:-8025}"
echo "   Hermes: http://localhost:${HERMES_PORT:-8081}"
echo ""
echo "🛑 To stop services:"
echo "   make dev-down"
echo ""

go run cmd/sphinx/main.go
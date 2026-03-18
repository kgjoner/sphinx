#!/bin/bash
set -e

echo "🚀 Setting up E2E test environment..."

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

# Start docker-compose services
docker-compose -f docker-compose.test.yaml up -d

# Wait for PostgreSQL to be ready
echo -e "\nWaiting for PostgreSQL to be ready..."
for i in {1..30}; do
	if docker exec sphinx-pg-test pg_isready -U postgres > /dev/null 2>&1; then
		echo -e "✅ PostgreSQL is ready!\n"
		break
	fi
	if [ $i -eq 30 ]; then
		echo -e "❌ PostgreSQL failed to start in time\n"
		docker-compose -f docker-compose.test.yaml logs pg-test
		exit 1
	fi
	sleep 1
done

# Wait for Redis to be ready
echo "Waiting for Redis to be ready..."
for i in {1..30}; do
	if docker exec sphinx-redis-test redis-cli ping > /dev/null 2>&1; then
		echo -e "✅ Redis is ready!\n"
		break
	fi
	if [ $i -eq 30 ]; then
		echo -e "❌ Redis failed to start in time\n"
		docker-compose -f docker-compose.test.yaml logs redis-test
		exit 1
	fi
	sleep 1
done

# Apply database migrations and seed test data
go run cmd/migrate/main.go
echo ""
go run cmd/testseed/main.go clean
go run cmd/testseed/main.go seed

echo ""
echo "✅ All services are ready!"
echo ""
echo "📝 Test environment details:"
echo "   PostgreSQL: localhost:${DB_PORT:-5432}"
echo "   Redis: localhost:${REDIS_PORT:-6379}"
echo "   Mailhog UI: http://localhost:${MAILHOG_WEB_PORT:-8025}"
echo "   Hermes: http://localhost:${HERMES_PORT:-8081}"
echo ""
echo "🧪 To run tests:"
echo "   make test-e2e"
echo ""
echo "🛑 To stop services:"
echo "   make test-env-down"

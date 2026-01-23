#!/bin/bash
set -e

echo "🚀 Setting up E2E test environment..."

# Start docker-compose services
docker-compose -f docker-compose.test.yaml up -d

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
	if docker exec sphinx-pg-test pg_isready -U postgres > /dev/null 2>&1; then
		echo "✅ PostgreSQL is ready!"
		break
	fi
	if [ $i -eq 30 ]; then
		echo "❌ PostgreSQL failed to start in time"
		docker-compose -f docker-compose.test.yaml logs pg-test
		exit 1
	fi
	sleep 1
done

# Wait for Redis to be ready
echo "Waiting for Redis to be ready..."
for i in {1..30}; do
	if docker exec sphinx-redis-test redis-cli ping > /dev/null 2>&1; then
		echo "✅ Redis is ready!"
		break
	fi
	if [ $i -eq 30 ]; then
		echo "❌ Redis failed to start in time"
		docker-compose -f docker-compose.test.yaml logs redis-test
		exit 1
	fi
	sleep 1
done

# Set environment variables for test services
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/sphinx_test?sslmode=disable
export REDIS_URL=redis://localhost:6380/0
export HERMES_BASE_URL=http://localhost:8082/v1

# Apply database migrations and seed test data
go run cmd/migrate/main.go
go run cmd/testseed/main.go clean
go run cmd/testseed/main.go seed

echo "✅ All services are ready!"
echo ""
echo "📝 Test environment details:"
echo "   PostgreSQL: localhost:5433"
echo "   Redis: localhost:6380"
echo "   Mailhog UI: http://localhost:8026"
echo "   Hermes: http://localhost:8082"
echo ""
echo "🧪 To run tests:"
echo "   make test-e2e"
echo ""
echo "🛑 To stop services:"
echo "   make test-e2e-down"

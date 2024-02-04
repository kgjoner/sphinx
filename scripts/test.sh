#! /bin/bash

if [ "$(docker ps -a -q -f name=test)" ]; then
  if [ "`docker container inspect -f '{{.State.Running}}' test`" = "false" ]; then
    docker start test
    until docker logs --tail 1 test 2>&1 | grep -q "database system is ready to accept connections"; do
      echo "wait for database to initialize..."
    done;
  fi
else
  docker run -d --name test -p 5432:5432 \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_DB=sphinx postgres
  until docker logs --tail 1 test 2>&1 | grep -q "database system is ready to accept connections"; do
    echo "wait for database to initialize..."
  done;
fi
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/sphinx?sslmode=disable
migrate -path postgres/migrations \
  -database $DATABASE_URL up
go test ./cmd
migrate -path postgres/migrations \
  -database $DATABASE_URL down -all
# docker stop test
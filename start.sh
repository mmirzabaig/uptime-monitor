#!/bin/bash

sudo systemctl stop postgresql

set -e

echo "Starting PostgreSQL..."

docker start uptime-postgres > /dev/null 2>&1 || true

echo "Waiting for PostgreSQL..."

until docker exec uptime-postgres pg_isready -U postgres -d uptime_monitor > /dev/null 2>&1; do
sleep 1
done

echo "PostgreSQL is ready."

echo "Running migrations..."

migrate \
  -path ./migrations \
  -database "postgres://postgres:password@localhost:5432/uptime_monitor?sslmode=disable" \
  up


echo "Starting application..."

go run ./cmd/server

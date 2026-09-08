#!/bin/bash

set -e

export DATABASE_URL="postgres://postgres:password@localhost:5432/uptime_monitor?sslmode=disable"
export SERVER_ADDR=":8080"

echo "Starting PostgreSQL..."

docker start uptime-postgres > /dev/null 2>&1 || true

echo "Waiting for PostgreSQL..."

until docker exec uptime-postgres pg_isready -U postgres -d uptime_monitor > /dev/null 2>&1; do
sleep 1
done

## COMMAND TO DELETE AFTER REBOOT
# sudo systemctl stop postgresql

## COMMAND TO START CONTAINER
# docker start uptime-postgres

## COMMAND TO START POSTGRES FROM SCRATCH IF CONTAINER DOES NOT EXIST
# docker run --name uptime-postgres \
#   -e POSTGRES_USER=postgres \
#   -e POSTGRES_PASSWORD=password \
#   -e POSTGRES_DB=uptime_monitor \
#   -p 5432:5432 \
#   -d postgres:14.24-alpine3.23

echo "PostgreSQL is ready."

echo "Running migrations..."

migrate \
  -path ./migrations \
  -database "postgres://postgres:password@localhost:5432/uptime_monitor?sslmode=disable" \
  up

echo "Starting application..."

go run ./cmd/server

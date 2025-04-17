#!/bin/bash

# Load environment variables
set -a
source .env.localhost
set +a

# Database URL for migrations
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

# Command to run
CMD=$1

# Use migrate from Go bin
MIGRATE_CMD="$HOME/go/bin/migrate"

case $CMD in
  "up")
    echo "Running all pending migrations..."
    $MIGRATE_CMD -database "${DB_URL}" -path migrations_new up
    ;;
  "down")
    echo "Rolling back last migration..."
    $MIGRATE_CMD -database "${DB_URL}" -path migrations_new down 1
    ;;
  "create")
    if [ -z "$2" ]; then
      echo "Please provide a migration name"
      exit 1
    fi
    echo "Creating new migration..."
    $MIGRATE_CMD create -ext sql -dir migrations_new -seq "$2"
    ;;
  *)
    echo "Usage: $0 <up|down|create> [migration_name]"
    echo "  up              : Run all pending migrations"
    echo "  down            : Rollback last migration"
    echo "  create [name]   : Create a new migration"
    exit 1
    ;;
esac 
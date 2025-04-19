#!/bin/bash

# Default environment file
ENV_FILE=".env.localhost"

# Parse options
while getopts "e:" opt; do
  case $opt in
    e)
      ENV_FILE="$OPTARG"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

# Shift to remove processed options
shift $((OPTIND-1))

# Load environment variables
set -a
if [ -f "$ENV_FILE" ]; then
  echo "Using environment file: $ENV_FILE"
  source "$ENV_FILE"
else
  echo "Warning: Environment file $ENV_FILE not found"
  exit 1
fi
set +a

# Database URL for migrations
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Command to run
CMD=$1

# Use migrate from Go bin
MIGRATE_CMD="migrate"

case $CMD in
  "up")
    echo "Running all pending migrations..."
    $MIGRATE_CMD -database "${DB_URL}" -path migrations up
    ;;
  "down")
    echo "Rolling back last migration..."
    $MIGRATE_CMD -database "${DB_URL}" -path migrations down 1
    ;;
  "create")
    if [ -z "$2" ]; then
      echo "Please provide a migration name"
      exit 1
    fi
    echo "Creating new migration..."
    $MIGRATE_CMD create -ext sql -dir migrations -seq "$2"
    ;;
  *)
    echo "Usage: $0 [-e env_file] <up|down|create> [migration_name]"
    echo "Options:"
    echo "  -e env_file    : Specify environment file (default: .env.localhost)"
    echo "Commands:"
    echo "  up              : Run all pending migrations"
    echo "  down            : Rollback last migration"
    echo "  create [name]   : Create a new migration"
    exit 1
    ;;
esac 
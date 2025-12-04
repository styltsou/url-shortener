#!/bin/bash
set -e

# Load environment variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."
source scripts/load-env.sh

if [ -z "$POSTGRES_CONNECTION_STRING" ]; then
  echo "Error: POSTGRES_CONNECTION_STRING environment variable is not set"
  exit 1
fi

# Get the last applied migration version
LAST_VERSION=$(psql "$POSTGRES_CONNECTION_STRING" -tAc "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null)

if [ -z "$LAST_VERSION" ]; then
  echo "No migrations have been applied yet"
  exit 1
fi

# Find the corresponding down migration file
DOWN_FILE=$(ls migrations/${LAST_VERSION}_*.down.sql 2>/dev/null | head -1)
if [ -z "$DOWN_FILE" ] || [ ! -f "$DOWN_FILE" ]; then
  echo "Error: Down migration file not found for version $LAST_VERSION"
  exit 1
fi

echo "Rolling back migration version $LAST_VERSION: $DOWN_FILE"
psql "$POSTGRES_CONNECTION_STRING" -f "$DOWN_FILE" || exit 1

# Remove from schema_migrations table
psql "$POSTGRES_CONNECTION_STRING" -c "DELETE FROM schema_migrations WHERE version = '$LAST_VERSION';" || exit 1

echo "✓ Rollback completed successfully!"


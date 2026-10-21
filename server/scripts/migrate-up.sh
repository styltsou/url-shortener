#!/bin/bash
set -e

# Load environment variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."
source scripts/load-env.sh

if [ -z "$POSTGRES_CONNECTION_STRING" ]; then
  echo "Error: POSTGRES_CONNECTION_STRING environment variable is not set"
  echo "Please set it in your .env file or export it:"
  echo "  export POSTGRES_CONNECTION_STRING=postgres://user:password@localhost:5432/dbname?sslmode=disable"
  exit 1
fi

# Create schema_migrations table if it doesn't exist
psql "$POSTGRES_CONNECTION_STRING" -v ON_ERROR_STOP=1 -c "
  CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW()
  );
" || exit 1

echo "Checking for pending migrations..."
PENDING_COUNT=0

# Process migrations in order
for file in migrations/*.up.sql; do
  if [ -f "$file" ]; then
    # Extract version from filename (e.g., 000001_create_links_table.up.sql -> 000001)
    VERSION=$(basename "$file" | sed 's/^\([0-9]*\)_.*/\1/')
    
    # Check if migration has already been run
    EXISTS=$(psql "$POSTGRES_CONNECTION_STRING" -v ON_ERROR_STOP=1 -tAc "SELECT COUNT(*) FROM schema_migrations WHERE version = '$VERSION';")
    
    if [ "$EXISTS" = "0" ]; then
      echo "Running migration: $file (version: $VERSION)"
      # Apply the SQL and its version record in one transaction.
      psql "$POSTGRES_CONNECTION_STRING" -v ON_ERROR_STOP=1 --single-transaction -f "$file" \
        -c "INSERT INTO schema_migrations (version) VALUES ('$VERSION');"
      
      PENDING_COUNT=$((PENDING_COUNT + 1))
      echo "✓ Migration $VERSION applied successfully"
    else
      echo "⊘ Migration $VERSION already applied, skipping"
    fi
  fi
done

if [ "$PENDING_COUNT" = "0" ]; then
  echo "No pending migrations. Database is up to date!"
else
  echo "Applied $PENDING_COUNT migration(s) successfully!"
fi


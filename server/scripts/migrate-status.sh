#!/bin/bash
set -e

# Load environment variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."
source scripts/load-env.sh

echo "Migration Status:"
echo "================="
echo ""

if [ -n "$POSTGRES_CONNECTION_STRING" ]; then
  # Check if schema_migrations table exists
  TABLE_EXISTS=$(psql "$POSTGRES_CONNECTION_STRING" -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations');" 2>/dev/null || echo "false")
  
  if [ "$TABLE_EXISTS" = "t" ]; then
    echo "Applied migrations:"
    psql "$POSTGRES_CONNECTION_STRING" -c "SELECT version, applied_at FROM schema_migrations ORDER BY version;" 2>/dev/null || echo "  (unable to query)"
    echo ""
  else
    echo "No migrations have been applied yet (schema_migrations table doesn't exist)"
    echo ""
  fi
else
  echo "Note: POSTGRES_CONNECTION_STRING not set, showing file list only"
  echo ""
fi

echo "Available migration files:"
echo ""
echo "Up migrations:"
ls -1 migrations/*.up.sql 2>/dev/null | sort -V | nl || echo "  No up migrations found"
echo ""
echo "Down migrations:"
ls -1 migrations/*.down.sql 2>/dev/null | sort -V | nl || echo "  No down migrations found"


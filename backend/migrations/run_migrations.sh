#!/bin/bash

# Full-Text Search Migration Script
# This script runs the PostgreSQL full-text search migration

set -e

echo "🔍 Running PostgreSQL Full-Text Search Migration..."

# Check if running in Docker environment
if command -v docker &> /dev/null && docker ps | grep -q incidex-db; then
    echo "📦 Docker environment detected"
    echo "Executing migration via Docker..."
    docker exec -i incidex-db psql -U postgres -d incidex < migrations/001_add_fulltext_search.sql
    echo "✅ Migration completed successfully!"
else
    echo "💻 Local PostgreSQL environment"
    echo "Please ensure PostgreSQL is running and accessible"
    echo ""
    read -p "Enter database host (default: localhost): " DB_HOST
    DB_HOST=${DB_HOST:-localhost}

    read -p "Enter database port (default: 5432): " DB_PORT
    DB_PORT=${DB_PORT:-5432}

    read -p "Enter database name (default: incidex): " DB_NAME
    DB_NAME=${DB_NAME:-incidex}

    read -p "Enter database user (default: postgres): " DB_USER
    DB_USER=${DB_USER:-postgres}

    echo ""
    echo "Executing migration..."
    PGPASSWORD="" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" < migrations/001_add_fulltext_search.sql
    echo "✅ Migration completed successfully!"
fi

echo ""
echo "📝 Verification:"
echo "You can verify the migration by running:"
echo "  SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'incidents' AND column_name = 'search_vector';"

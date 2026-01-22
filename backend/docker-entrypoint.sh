#!/bin/sh
set -e

# Run database migrations if AUTO_MIGRATE is enabled
if [ "$AUTO_MIGRATE" = "true" ]; then
    echo "INFO: AUTO_MIGRATE is enabled. Running database migrations..."

    # Extract database connection parameters from DATABASE_URL
    # Wait for database to be ready
    until goose -dir "${MIGRATIONS_DIR:-/root/migrations}" postgres "$DATABASE_URL" version > /dev/null 2>&1; do
        echo "Waiting for database to be ready..."
        sleep 2
    done

    # Run migrations
    echo "Running goose migrations..."
    goose -dir "${MIGRATIONS_DIR:-/root/migrations}" postgres "$DATABASE_URL" up

    echo "SUCCESS: Database migrations completed"
else
    echo "INFO: AUTO_MIGRATE is disabled. Skipping migrations."
fi

# Execute the main application
echo "Starting application..."
exec "$@"

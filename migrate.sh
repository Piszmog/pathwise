#!/bin/bash
set -e

show_usage() {
    echo "Usage: $0 -p <protocol> -u <database_url> [-d <direction>] [-t <auth_token>] [-s <steps>]"
    echo ""
    echo "Flags:"
    echo "  -p, --protocol     Database protocol (required: sqlite3, libsql, postgres, etc.)"
    echo "  -u, --url          Database URL without protocol (required)"
    echo "  -d, --direction    Migration direction: up (default) or down"
    echo "  -t, --token        Authentication token for remote databases"
    echo "  -s, --steps        Number of steps for down migration (default: 1)"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -p sqlite3 -u ./db.sqlite3"
    echo "  $0 -p sqlite3 -u ./db.sqlite3 -d up"
    echo "  $0 -p libsql -u pathwise-local-piszmog.aws-us-west-2.turso.io -d up -t your_token"
    echo "  $0 --protocol postgres --url localhost:5432/mydb --direction down --steps 3"
}

ensure_migrate_tools() {
    echo "Installing migration tools..."
    go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    go install github.com/Piszmog/migrate-libsql@latest
}

# Initialize variables
PROTOCOL=""
DB_URL=""
DIRECTION="up"
AUTH_TOKEN=""
STEPS="1"


# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--protocol)
            PROTOCOL="$2"
            shift 2
            ;;
        -u|--url)
            DB_URL="$2"
            shift 2
            ;;
        -d|--direction)
            DIRECTION="$2"
            shift 2
            ;;
        -t|--token)
            AUTH_TOKEN="$2"
            shift 2
            ;;
        -s|--steps)
            STEPS="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "Error: Unknown option $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate required arguments
if [ -z "$PROTOCOL" ]; then
    echo "Error: Protocol is required (use -p or --protocol)"
    show_usage
    exit 1
fi

if [ -z "$DB_URL" ]; then
    echo "Error: Database URL is required (use -u or --url)"
    show_usage
    exit 1
fi

# Validate direction
if [[ "$DIRECTION" != "up" && "$DIRECTION" != "down" ]]; then
    echo "Error: Direction must be 'up' or 'down'"
    show_usage
    exit 1
fi

# Construct full database URL with optional token
if [ -n "$AUTH_TOKEN" ]; then
    FULL_DB_URL="${PROTOCOL}://${DB_URL}?authToken=${AUTH_TOKEN}"
else
    FULL_DB_URL="${PROTOCOL}://${DB_URL}"
fi

run_migration() {
    echo "Running migration $DIRECTION with $PROTOCOL://$DB_URL"

    if [ "$PROTOCOL" == "libsql" ]; then
        if [ "$DIRECTION" == "down" ]; then
            echo "Running libsql down1"
            migrate-libsql \
                -url "$PROTOCOL://$DB_URL" \
                -token "$AUTH_TOKEN" \
                -migrations ./internal/db/migrations \
                -direction down \
                -steps "$STEPS"
        else
            echo "Running libsql up2"
            migrate-libsql \
                -url "$PROTOCOL://$DB_URL" \
                -token "$AUTH_TOKEN" \
                -migrations ./internal/db/migrations \
                -direction up
        fi
    else
        if [ "$DIRECTION" == "down" ]; then
            echo "Running sqlite3 up3"
            migrate \
                -source file://./internal/db/migrations \
                -database "$FULL_DB_URL" \
                down "$STEPS"
        else
            echo "Running sqlite3 up4"
            migrate \
                -source file://./internal/db/migrations \
                -database "$FULL_DB_URL" \
                up
        fi
    fi

    echo "Migration completed successfully"
}

echo "Starting database migration..."
ensure_migrate_tools
run_migration

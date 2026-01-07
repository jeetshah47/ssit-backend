#!/bin/bash
# Bash script to truncate all tables in the database
# Usage: ./truncate-tables.sh [--confirm] [--tables "users,otps"] [--verify] [--list]

set -e

# Default values
CONFIRM=false
TABLES=""
VERIFY=false
LIST=false
ENV_FILE=".env"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --confirm)
            CONFIRM=true
            shift
            ;;
        --tables)
            TABLES="$2"
            shift 2
            ;;
        --verify)
            VERIFY=true
            shift
            ;;
        --list)
            LIST=true
            shift
            ;;
        --env)
            ENV_FILE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--confirm] [--tables \"users,otps\"] [--verify] [--list] [--env .env]"
            exit 1
            ;;
    esac
done

# Load environment variables from .env file if it exists
if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

# Get database connection parameters
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-equitywala}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# Build connection string
if [ -n "$DATABASE_URL" ]; then
    DB_URL="$DATABASE_URL"
else
    DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
fi

# Define all tables in truncation order (child tables first)
ALL_TABLES=(
    "payment_webhooks"
    "payments"
    "voucher_redemptions"
    "subscriptions"
    "vouchers"
    "pricing_packages"
    "advisory_views"
    "advisories"
    "team_members"
    "notification_settings"
    "user_classification_logs"
    "kyc_records"
    "user_roles"
    "sessions"
    "otps"
    "roles"
    "users"
)

# Determine which tables to truncate
if [ -n "$TABLES" ]; then
    IFS=',' read -ra TABLES_TO_TRUNCATE <<< "$TABLES"
    for i in "${!TABLES_TO_TRUNCATE[@]}"; do
        TABLES_TO_TRUNCATE[$i]=$(echo "${TABLES_TO_TRUNCATE[$i]}" | xargs)
    done
else
    TABLES_TO_TRUNCATE=("${ALL_TABLES[@]}")
fi

# List tables if requested
if [ "$LIST" = true ]; then
    echo "Tables that would be truncated:"
    for table in "${TABLES_TO_TRUNCATE[@]}"; do
        count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs)
        echo "  - $table ($count rows)"
    done
    exit 0
fi

# Safety check
if [ "$CONFIRM" != true ]; then
    echo "ERROR: Truncation requires confirmation flag"
    echo "Usage: $0 --confirm"
    echo "Or: $0 --confirm --tables 'users,otps'"
    echo "Use --list to see tables and row counts"
    exit 1
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "ERROR: psql command not found. Please install PostgreSQL client tools."
    echo "On Ubuntu/Debian: sudo apt-get install postgresql-client"
    echo "On macOS: brew install postgresql"
    exit 1
fi

# Create SQL script
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

cat > "$TEMP_FILE" <<EOF
BEGIN;

EOF

for table in "${TABLES_TO_TRUNCATE[@]}"; do
    echo "TRUNCATE TABLE $table CASCADE;" >> "$TEMP_FILE"
done

cat >> "$TEMP_FILE" <<EOF

COMMIT;
EOF

# Execute SQL script
echo "Truncating tables..."
if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$TEMP_FILE" > /dev/null 2>&1; then
    echo ""
    echo "✓ All tables truncated successfully!"
else
    echo ""
    echo "✗ Error truncating tables"
    exit 1
fi

# Verify if requested
if [ "$VERIFY" = true ]; then
    echo ""
    echo "Verifying tables are empty..."
    all_empty=true
    for table in "${TABLES_TO_TRUNCATE[@]}"; do
        count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs)
        if [ "$count" = "0" ] || [ -z "$count" ]; then
            echo "  ✓ $table : empty"
        else
            echo "  ✗ $table : $count rows (not empty!)"
            all_empty=false
        fi
    done
    if [ "$all_empty" = true ]; then
        echo ""
        echo "✓ All tables verified as empty!"
    else
        echo ""
        echo "✗ Some tables are not empty!"
        exit 1
    fi
fi


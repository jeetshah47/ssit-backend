# Database Migrations

This directory contains database migration scripts for the Equitywala Stock Advisory Platform.

## Migration Files

### PostgreSQL Migrations

**Core Schema:**
- `001_initial_schema.up.sql` - Creates all initial tables
- `001_initial_schema.down.sql` - Rolls back initial schema
- `002_add_user_profile_fields.up.sql` - Adds profile fields (DOB, City, PAN) and makes password_hash nullable
- `002_add_user_profile_fields.down.sql` - Rolls back profile fields migration
- `003_add_customer_type_field.up.sql` - Adds customer_type field to users table
- `003_add_customer_type_field.down.sql` - Rolls back customer type field

**Payment & Pricing:**
- `004_add_payment_plan_fields.up.sql` - Adds payment plan fields
- `004_add_payment_plan_fields.down.sql` - Rolls back payment plan fields
- `005_create_user_payment_plan_selections.up.sql` - Creates user payment plan selections table
- `005_create_user_payment_plan_selections.down.sql` - Rolls back payment plan selections table
- `006_remove_payment_plan_fields_from_users.up.sql` - Removes payment plan fields from users
- `006_remove_payment_plan_fields_from_users.down.sql` - Rolls back removal
- `007_remove_discount_voucher_fields.up.sql` - Removes discount and voucher fields
- `007_remove_discount_voucher_fields.down.sql` - Rolls back removal
- `008_seed_pricing_packages.up.sql` - Seeds pricing packages
- `008_seed_pricing_packages.down.sql` - Rolls back pricing packages seed
- `009_update_payment_method_constraint.up.sql` - Updates payment method constraint
- `009_update_payment_method_constraint.down.sql` - Rolls back constraint update
- `010_update_payment_method_for_paytm.up.sql` - Updates payment method for Paytm
- `010_update_payment_method_for_paytm.down.sql` - Rolls back Paytm update
- `011_fix_payment_method_constraint.up.sql` - Fixes payment method constraint
- `011_fix_payment_method_constraint.down.sql` - Rolls back constraint fix

**Advisory Content:**
- `012_create_advisory_types.up.sql` - Creates advisory types table
- `012_create_advisory_types.down.sql` - Rolls back advisory types
- `013_create_stock_baskets.up.sql` - Creates stock baskets and items tables
- `013_create_stock_baskets.down.sql` - Rolls back stock baskets
- `014_create_etf_baskets.up.sql` - Creates ETF baskets and items tables
- `014_create_etf_baskets.down.sql` - Rolls back ETF baskets
- `015_create_ipo_advisories.up.sql` - Creates IPO advisories table
- `015_create_ipo_advisories.down.sql` - Rolls back IPO advisories
- `016_create_mutual_fund_baskets.up.sql` - Creates mutual fund baskets and items tables
- `016_create_mutual_fund_baskets.down.sql` - Rolls back mutual fund baskets
- `017_create_sector_snapshots.up.sql` - Creates sector snapshots table
- `017_create_sector_snapshots.down.sql` - Rolls back sector snapshots
- `018_create_mf_schemes.up.sql` - Creates MF schemes table
- `018_create_mf_schemes.down.sql` - Rolls back MF schemes
- `019_create_nfos.up.sql` - Creates NFOs (New Fund Offers) table
- `019_create_nfos.down.sql` - Rolls back NFOs
- `020_create_webinars.up.sql` - Creates webinars table
- `020_create_webinars.down.sql` - Rolls back webinars
- `021_create_weekly_market_mood.up.sql` - Creates weekly market mood table
- `021_create_weekly_market_mood.down.sql` - Rolls back weekly market mood
- `022_create_weekly_audio.up.sql` - Creates weekly audio table
- `022_create_weekly_audio.down.sql` - Rolls back weekly audio
- `023_seed_dummy_data.up.sql` - Seeds dummy data for development
- `023_seed_dummy_data.down.sql` - Rolls back dummy data
- `024_add_slow_query_indexes.up.sql` - Adds indexes for ipo_advisories (created_at) and stock_baskets (status, is_bullet_idea, published_at)
- `024_add_slow_query_indexes.down.sql` - Rolls back slow-query indexes
- `025_create_stocks.up.sql` - Creates master stocks table (for bullets/recommendations selection)
- `025_create_stocks.down.sql` - Rolls back stocks table
- `026_add_stock_id_to_stock_basket_items.up.sql` - Adds stock_id FK to stock_basket_items, backfills from stocks, drops stock_name/stock_symbol
- `026_add_stock_id_to_stock_basket_items.down.sql` - Restores stock_name/stock_symbol, drops stock_id

## Running Migrations

### Using Go Migration Script (Recommended)

The project includes a Go script for running migrations that reads database configuration from environment variables.

**Build the migration script:**
```bash
cd ssit-backend
go build -o scripts/migrate.exe ./scripts/migrate.go
```

**Run migrations:**

The script uses environment variables for database configuration (same as the main application):
- `DB_HOST` (default: localhost)
- `DB_PORT` (default: 5432)
- `DB_USER` (default: postgres)
- `DB_PASSWORD` (default: postgres)
- `DB_NAME` (default: equitywala)
- `DB_SSLMODE` (default: disable)

**Examples:**

```bash
# Apply all pending migrations (from project root)
cd ssit-backend
go run ./scripts/migrate.go -command up

# Or use the compiled binary
./scripts/migrate.exe -command up

# Rollback all migrations
go run ./scripts/migrate.go -command down

# Apply specific number of migrations
go run ./scripts/migrate.go -command up -steps 2

# Rollback specific number of migrations
go run ./scripts/migrate.go -command down -steps 1

# Migrate to specific version
go run ./scripts/migrate.go -command goto -version 3

# Check current migration version
go run ./scripts/migrate.go -command version

# Force migration version (use with caution)
go run ./scripts/migrate.go -command force -version 2

# Override database connection (if not using environment variables)
go run ./scripts/migrate.go -command up -host localhost -port 5432 -user postgres -password postgres -dbname equitywala
```

### Using golang-migrate CLI (Alternative)

Install golang-migrate:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate

# Windows
# Download from https://github.com/golang-migrate/migrate/releases
```

Run migrations:
```bash
# From ssit-backend directory
# Up migration
migrate -path database/migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" up

# Down migration (rollback)
migrate -path database/migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" down

# To specific version
migrate -path database/migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" goto 1
```

### Using psql

```bash
# Run up migration
psql -U postgres -d equitywala -f migrations/postgres/001_initial_schema.up.sql

# Run down migration (rollback)
psql -U postgres -d equitywala -f migrations/postgres/001_initial_schema.down.sql
```

### Using Docker

If using docker-compose:
```bash
docker-compose exec postgres psql -U postgres -d equitywala -f /migrations/001_initial_schema.up.sql
```

## Migration Naming Convention

- `{version}_{description}.up.sql` - Forward migration
- `{version}_{description}.down.sql` - Rollback migration

Example:
- `001_initial_schema.up.sql`
- `001_initial_schema.down.sql`
- `002_add_user_profile_fields.up.sql`
- `002_add_user_profile_fields.down.sql`

## Creating New Migrations

1. Create new migration files following the naming convention
2. Test migrations on development database first
3. Ensure down migration properly reverses up migration
4. Update this README if needed

## Notes

- Always test migrations on a development database first
- Never modify existing migration files that have been run in production
- Create new migrations for schema changes
- Keep migrations small and focused


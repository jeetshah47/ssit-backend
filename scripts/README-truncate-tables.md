# Table Truncation Scripts

This directory contains scripts to truncate (empty) all tables in the PostgreSQL database. These scripts are useful for:
- Resetting the database during development
- Clearing test data
- Starting fresh with migrations

## ⚠️ WARNING

**These scripts will DELETE ALL DATA from the specified tables. Use with extreme caution!**

Always backup your database before running these scripts in production or with important data.

## Available Scripts

### 1. SQL Script (`truncate-tables.sql`)

Direct SQL script that can be executed with `psql`.

**Usage:**
```bash
# Using connection string
psql $DATABASE_URL -f truncate-tables.sql

# Using individual parameters
psql -h localhost -p 5432 -U postgres -d equitywala -f truncate-tables.sql
```

### 2. Go Script (`truncate-tables.go`)

Go program that provides additional safety checks and features.

**Prerequisites:**
- Go installed
- PostgreSQL driver: `go get github.com/lib/pq`
- Environment variables or `.env` file configured

**Usage:**
```bash
# List tables and row counts (safe, read-only)
go run truncate-tables.go -list

# Truncate all tables (requires -confirm flag)
go run truncate-tables.go -confirm

# Truncate specific tables
go run truncate-tables.go -confirm -tables=users,otps,sessions

# Truncate and verify
go run truncate-tables.go -confirm -verify

# Use custom .env file
go run truncate-tables.go -confirm -env=.env.local
```

**Options:**
- `-confirm`: Required flag to confirm truncation (safety measure)
- `-tables`: Comma-separated list of specific tables to truncate (default: all tables)
- `-verify`: Verify tables are empty after truncation
- `-list`: List all tables with row counts (read-only, safe)
- `-env`: Path to .env file (default: `.env`)

### 3. PowerShell Script (`truncate-tables.ps1`)

Windows PowerShell script for Windows users.

**Prerequisites:**
- PostgreSQL client tools (`psql`) installed
- Environment variables or `.env` file configured

**Usage:**
```powershell
# List tables and row counts (safe, read-only)
.\truncate-tables.ps1 -List

# Truncate all tables (requires -Confirm flag)
.\truncate-tables.ps1 -Confirm

# Truncate specific tables
.\truncate-tables.ps1 -Confirm -Tables "users,otps,sessions"

# Truncate and verify
.\truncate-tables.ps1 -Confirm -Verify

# Use custom .env file
.\truncate-tables.ps1 -Confirm -EnvFile ".env.local"
```

**Options:**
- `-Confirm`: Required switch to confirm truncation (safety measure)
- `-Tables`: Comma-separated list of specific tables to truncate (default: all tables)
- `-Verify`: Verify tables are empty after truncation
- `-List`: List all tables with row counts (read-only, safe)
- `-EnvFile`: Path to .env file (default: `.env`)

### 4. Bash Script (`truncate-tables.sh`)

Unix/Linux/macOS bash script.

**Prerequisites:**
- PostgreSQL client tools (`psql`) installed
- Environment variables or `.env` file configured
- Execute permission: `chmod +x truncate-tables.sh`

**Usage:**
```bash
# List tables and row counts (safe, read-only)
./truncate-tables.sh --list

# Truncate all tables (requires --confirm flag)
./truncate-tables.sh --confirm

# Truncate specific tables
./truncate-tables.sh --confirm --tables "users,otps,sessions"

# Truncate and verify
./truncate-tables.sh --confirm --verify

# Use custom .env file
./truncate-tables.sh --confirm --env .env.local
```

**Options:**
- `--confirm`: Required flag to confirm truncation (safety measure)
- `--tables`: Comma-separated list of specific tables to truncate (default: all tables)
- `--verify`: Verify tables are empty after truncation
- `--list`: List all tables with row counts (read-only, safe)
- `--env`: Path to .env file (default: `.env`)

## Environment Variables

All scripts support the following environment variables (can be set in `.env` file):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=equitywala
DB_SSLMODE=disable

# Or use a connection string
DATABASE_URL=postgresql://user:password@host:port/dbname?sslmode=disable
```

## Tables Truncated (in order)

The scripts truncate tables in the correct order to respect foreign key constraints:

1. `payment_webhooks`
2. `payments`
3. `voucher_redemptions`
4. `subscriptions`
5. `vouchers`
6. `pricing_packages`
7. `advisory_views`
8. `advisories`
9. `team_members`
10. `notification_settings`
11. `user_classification_logs`
12. `kyc_records`
13. `user_roles`
14. `sessions`
15. `otps`
16. `roles`
17. `users`

## Examples

### Development Reset

```bash
# Quick reset during development
go run truncate-tables.go -confirm

# Or with SQL
psql $DATABASE_URL -f truncate-tables.sql
```

### Selective Truncation

```bash
# Only clear auth-related tables
go run truncate-tables.go -confirm -tables=users,otps,sessions,user_roles

# Only clear payment data
go run truncate-tables.go -confirm -tables=payments,payment_webhooks,subscriptions
```

### Safe Inspection

```bash
# Check what would be deleted (read-only)
go run truncate-tables.go -list
```

## Notes

- All scripts use `TRUNCATE ... CASCADE` to handle foreign key constraints automatically
- Tables are truncated in a transaction for atomicity
- The Go script provides the most features and safety checks
- The SQL script is the simplest and most portable
- PowerShell and Bash scripts require `psql` to be in your PATH

## Troubleshooting

### "psql: command not found"
Install PostgreSQL client tools:
- **Windows**: Download from [PostgreSQL Downloads](https://www.postgresql.org/download/windows/)
- **macOS**: `brew install postgresql`
- **Ubuntu/Debian**: `sudo apt-get install postgresql-client`

### "Connection refused" or authentication errors
- Check your database connection settings in `.env` or environment variables
- Verify the database is running
- Check firewall settings
- Verify credentials are correct

### "Permission denied" (bash script)
Make the script executable:
```bash
chmod +x truncate-tables.sh
```

### Go script dependencies
Install required Go packages:
```bash
go get github.com/lib/pq
go get github.com/joho/godotenv
```


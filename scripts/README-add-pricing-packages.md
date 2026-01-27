# Add Pricing Packages Script

This script adds pricing packages to the `pricing_packages` table in the database.

## Usage

```bash
# Basic usage (inserts new packages, skips existing ones)
go run scripts/add-pricing-packages.go

# Force update existing packages
go run scripts/add-pricing-packages.go -force

# Clear all existing packages before inserting
go run scripts/add-pricing-packages.go -clear

# Specify custom .env file
go run scripts/add-pricing-packages.go -env=.env.local
```

## Flags

- `-env`: Path to .env file (default: `.env`)
- `-force`: Update existing packages if they already exist
- `-clear`: Delete all existing packages before inserting new ones

## Environment Variables

The script uses the same database connection variables as other scripts:

- `DATABASE_URL`: Full database connection string (takes precedence)
- `DB_HOST`: Database host (default: `localhost`)
- `DB_PORT`: Database port (default: `5432`)
- `DB_USER`: Database user (default: `postgres`)
- `DB_PASSWORD`: Database password (default: `postgres`)
- `DB_NAME`: Database name (default: `equitywala`)
- `DB_SSLMODE`: SSL mode (default: `disable`)

## Packages Created

The script creates the following pricing packages:

### Standard Plan
- **Quarterly**: ₹249.75 (90 days)
- **Annual**: ₹999.00 (365 days)
- **Access Level**: standard
- **Features**: Basic advisory, monthly reports, email support

### Plus Plan (Recommended)
- **Quarterly**: ₹499.75 (90 days)
- **Annual**: ₹1,999.00 (365 days)
- **Access Level**: plus
- **Features**: Premium advisory, weekly reports, priority support, stock baskets

### Premium Plan
- **Quarterly**: ₹749.75 (90 days)
- **Annual**: ₹2,999.00 (365 days)
- **Access Level**: premium
- **Features**: All Plus features, daily reports, 24/7 support, personal advisor, IPO access

All packages are set to:
- **Status**: `active`
- **Is Published**: `true`
- **Currency**: `INR`

## Examples

```bash
# Insert packages (skip if exists)
go run scripts/add-pricing-packages.go

# Update existing packages
go run scripts/add-pricing-packages.go -force

# Start fresh (clear all and insert)
go run scripts/add-pricing-packages.go -clear
```

## Notes

- The script uses transactions to ensure data consistency
- Existing packages are identified by name and duration_type combination
- Use `-force` flag to update existing packages with new values
- Use `-clear` flag to remove all existing packages before inserting

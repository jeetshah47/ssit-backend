# Database Migrations

This directory contains database migration scripts for the Equitywala Stock Advisory Platform.

## Migration Files

### PostgreSQL Migrations

- `001_initial_schema.up.sql` - Creates all initial tables
- `001_initial_schema.down.sql` - Rolls back initial schema
- `002_add_user_profile_fields.up.sql` - Adds profile fields (DOB, City, PAN) and makes password_hash nullable
- `002_add_user_profile_fields.down.sql` - Rolls back profile fields migration

## Running Migrations

### Using golang-migrate (Recommended)

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
# Up migration
migrate -path migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" up

# Down migration (rollback)
migrate -path migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" down

# To specific version
migrate -path migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" goto 1
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


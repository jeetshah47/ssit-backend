# Fixing Dirty Database State

When a migration fails partway through, the database is marked as "dirty" to prevent further migrations until the issue is resolved.

## Error Message
```
error: Dirty database version 1. Fix and force version.
```

## Solution

### Step 1: Check Current Migration Version
```bash
migrate -path migrations/postgres -database "postgres://postgres:postgres@localhost:5432/equitywala?sslmode=disable" version
```

### Step 2: Check What Actually Happened
Connect to your database and check if the tables were created:
```bash
psql -U postgres -d equitywala -c "\dt"
```

### Step 3: Fix the Dirty State

**Option A: If migration 1 partially succeeded (some tables created)**
Force the version to 1 (marking it as complete):
```bash
migrate -path migrations/postgres -database "postgres://postgres:postgres@localhost:5432/equitywala?sslmode=disable" force 1
```

**Option B: If migration 1 completely failed (no tables created)**
Force the version to 0 (rollback to clean state):
```bash
migrate -path migrations/postgres -database "postgres://postgres:postgres@localhost:5432/equitywala?sslmode=disable" force 0
```

### Step 4: Re-run Migrations

After fixing the dirty state, run migrations again:
```bash
migrate -path migrations/postgres -database "postgres://postgres:postgres@localhost:5432/equitywala?sslmode=disable" up
```

## Alternative: Manual Cleanup

If the above doesn't work, you can manually clean up:

1. **Drop and recreate the database:**
```bash
psql -U postgres -c "DROP DATABASE IF EXISTS equitywala;"
psql -U postgres -c "CREATE DATABASE equitywala;"
```

2. **Then run migrations fresh:**
```bash
migrate -path migrations/postgres -database "postgres://postgres:postgres@localhost:5432/equitywala?sslmode=disable" up
```

## Prevention

To avoid dirty states in the future:
- Always test migrations on a development database first
- Use transactions where possible
- Check for syntax errors before running migrations
- Use `IF NOT EXISTS` clauses where appropriate


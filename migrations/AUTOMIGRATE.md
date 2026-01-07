# GORM AutoMigrate

This project uses GORM's AutoMigrate feature for database schema management during development.

## How It Works

GORM AutoMigrate automatically creates/updates database tables based on the GORM model structs defined in the codebase. When the application starts, it:

1. Connects to the PostgreSQL database
2. Enables the UUID extension (if not already enabled)
3. Runs AutoMigrate on all registered models
4. Creates or updates tables to match the model definitions

## Models Currently Migrated

### Auth Module
- `users` - User accounts and profiles
- `otps` - OTP codes for email verification
- `sessions` - Active login sessions

### Future Modules (To Be Added)
- KYC models
- Subscription models
- Payment models
- Advisory models
- Settings models
- Admin models
- Roles models

## Usage

AutoMigrate runs automatically when the application starts. No manual migration commands are needed.

```bash
# Just start the server
make run
# or
go run ./cmd/server
```

The migration happens in `cmd/server/main.go` after database connection:

```go
// Run AutoMigrate
if err := postgres.AutoMigrate(postgresDB); err != nil {
    appLogger.Fatal("Failed to run database migrations", "error", err)
}
```

## Adding New Models

To add a new model for AutoMigrate:

1. Create the GORM model struct in the appropriate package (e.g., `internal/infrastructure/database/postgres/{module}/model.go`)
2. Add the model to the `AutoMigrate` call in `internal/infrastructure/database/postgres/migrate.go`:

```go
if err := db.AutoMigrate(
    &user.UserModel{},
    &auth.OTPModel{},
    &auth.SessionModel{},
    &newmodule.NewModel{}, // Add your new model here
); err != nil {
    return fmt.Errorf("failed to migrate: %w", err)
}
```

3. Restart the application - the table will be created automatically

## GORM Model Tags

Use GORM tags to define table structure:

```go
type UserModel struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
    DeletedAt *time.Time `gorm:"index"` // For soft deletes
}
```

Common GORM tags:
- `type:uuid` - PostgreSQL UUID type
- `primary_key` - Primary key
- `uniqueIndex` - Unique index
- `index` - Regular index
- `not null` - NOT NULL constraint
- `default:value` - Default value
- `autoCreateTime` - Auto-set on create
- `autoUpdateTime` - Auto-update on update

## Limitations

### What AutoMigrate Does
- ✅ Creates tables
- ✅ Adds missing columns
- ✅ Creates indexes
- ✅ Adds foreign keys (with proper setup)

### What AutoMigrate Does NOT Do
- ❌ Remove unused columns
- ❌ Change column types (may cause data loss)
- ❌ Remove indexes
- ❌ Handle complex migrations (data transformations, etc.)

## Production Considerations

For production environments, consider:

1. **Manual Migrations**: Use SQL migration files for production deployments
2. **Migration Tools**: Use tools like `golang-migrate` for version-controlled migrations
3. **Review Changes**: Always review what AutoMigrate will change before running in production
4. **Backup First**: Always backup production database before migrations

## Manual SQL Migrations

If you need more control, SQL migration files are available in `migrations/postgres/`:
- `001_initial_schema.up.sql` - Creates all tables
- `001_initial_schema.down.sql` - Rolls back schema

These can be used with migration tools like `golang-migrate` for production deployments.

## Troubleshooting

### UUID Extension Error
If you see an error about UUID extension:
```sql
-- Run this manually in PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Column Type Changes
If you need to change a column type, you may need to:
1. Create a manual migration SQL file
2. Or drop and recreate the table (⚠️ data loss)

### Index Conflicts
If AutoMigrate fails due to index conflicts:
1. Check existing indexes: `\d table_name` in psql
2. Drop conflicting indexes manually
3. Restart the application

## Best Practices

1. **Development**: Use AutoMigrate for rapid iteration
2. **Staging**: Test AutoMigrate changes before production
3. **Production**: Use manual SQL migrations for controlled deployments
4. **Version Control**: Keep model changes in version control
5. **Documentation**: Document schema changes in commit messages

---

**Note**: AutoMigrate is convenient for development but should be used carefully in production. Always test migrations in a staging environment first.


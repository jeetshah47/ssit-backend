# Scripts Directory

This directory contains **essential operational scripts** for the backend.

## Essential Scripts

### `migrate.go`
Database migration tool for PostgreSQL.

**Usage:**
```powershell
# Run migrations
go run scripts/migrate.go -command=up

# Rollback migrations
go run scripts/migrate.go -command=down

# Check current version
go run scripts/migrate.go -command=version
```

**Environment Variables:**
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (default: equitywala)
- `DB_SSLMODE` - SSL mode (default: disable)

**Custom .env file:**
```powershell
go run scripts/migrate.go -command=up -env=.env.sandbox
```

## Other Essential Scripts (Root Directory)

### `build-docker.ps1` / `build-docker.sh`
Complete Docker build and ECR push workflow.

**Usage:**
```powershell
# Build and push to ECR
.\build-docker.ps1 -ImageName equitywala-server -Tag latest -EcrRepoUrl 128977214845.dkr.ecr.ap-south-1.amazonaws.com/equitywala-server
```

### `aws-ecs/update-task-definition.ps1`
Update ECS task definition and service.

**Usage:**
```powershell
cd aws-ecs
.\update-task-definition.ps1 -ClusterName equitywala-dev-server
```

## Utility Scripts

One-time setup and maintenance scripts have been moved to `scripts/utils/`:

- `add-admin-user.go` - Create admin user (one-time setup)
- `add-pricing-packages.go` - Add pricing packages (one-time setup)
- `truncate-tables.go` - Database maintenance
- `drop-tables.go` - Database maintenance
- `fix-payment-constraint.go` - One-time fix script

See `scripts/utils/` for documentation on these scripts.

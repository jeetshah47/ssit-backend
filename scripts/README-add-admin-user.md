# Add Admin User Script

This script creates an admin user in the database with the specified role.

## Usage

```bash
# Basic usage (requires email, name, and password)
go run scripts/add-admin-user.go -email=admin@example.com -name="Admin User" -password=securepassword123

# With custom role
go run scripts/add-admin-user.go -email=admin@example.com -name="Admin User" -password=securepassword123 -role=super_admin

# Update existing user (force)
go run scripts/add-admin-user.go -email=admin@example.com -name="Admin User" -password=newpassword123 -force

# With custom .env file
go run scripts/add-admin-user.go -email=admin@example.com -name="Admin User" -password=securepassword123 -env=.env.local
```

## Parameters

- `-email` (required): Email address for the admin user
- `-name` (required): Full name of the admin user
- `-password` (required): Password for the admin user (will be hashed using bcrypt)
- `-role` (optional): Role name to assign (default: "admin")
- `-env` (optional): Path to .env file (default: ".env")
- `-force` (optional): Force creation/update even if user already exists

## Environment Variables

The script uses the same database connection variables as other scripts:

- `DATABASE_URL` (preferred) - Full PostgreSQL connection string
- Or individual variables:
  - `DB_HOST` (default: "localhost")
  - `DB_PORT` (default: "5432")
  - `DB_USER` (default: "postgres")
  - `DB_PASSWORD` (default: "postgres")
  - `DB_NAME` (default: "equitywala")
  - `DB_SSLMODE` (default: "disable")

## What the Script Does

1. Connects to the PostgreSQL database using environment variables
2. Creates a new user with:
   - Email, name, and hashed password
   - Status set to "active"
   - Email verified flag set to true
3. Creates the specified role if it doesn't exist (marks it as a system role)
4. Assigns the role to the user via the `user_roles` table
5. Updates the role's member count

## Notes

- If a user with the same email already exists, the script will fail unless `-force` is used
- When `-force` is used, the existing user's password, name, and status will be updated
- The role will be created automatically if it doesn't exist
- Passwords are hashed using bcrypt with default cost

## Example

```bash
# Create an admin user
go run scripts/add-admin-user.go \
  -email=admin@equitywala.com \
  -name="System Administrator" \
  -password=ChangeMe123! \
  -role=admin
```

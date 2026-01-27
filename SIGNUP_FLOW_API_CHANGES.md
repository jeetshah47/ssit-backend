# Signup Flow API Changes

This document describes the changes made to align the backend APIs with the frontend signup flow.

## Overview

The signup process has been changed from a single-step registration to a multi-step flow:

1. **Step 1**: Initial signup (name + email) → Send OTP
2. **Step 2**: Verify email with 4-digit OTP
3. **Step 3**: Update profile (mobile, DOB, city)
4. **Step 4**: Verify PAN
5. **Step 5**: Set password

## API Changes

### 1. Signup API (`POST /api/v1/auth/signup`)

**Changed:**
- Password is now **optional** (can be empty for step-by-step signup)
- Phone field removed from request

**Request:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "" // Optional - can be omitted or empty
}
```

**Response:**
```json
{
  "userId": "uuid",
  "email": "user@example.com",
  "requiresOTPVerification": true,
  "message": "OTP sent to your email address"
}
```

### 2. OTP Verification (`POST /api/v1/auth/verify-otp`)

**Changed:**
- OTP code changed from **6 digits to 4 digits**

**Request:**
```json
{
  "email": "user@example.com",
  "otp": "1234" // 4-digit code
}
```

**Response:**
- If password is not set: Returns user info without token
- If password is set: Returns token and user info

### 3. Update Profile (`PUT /api/v1/auth/profile`) - **NEW**

**Requires:** Authentication token

**Request:**
```json
{
  "phone": "1234567890",
  "dateOfBirth": "1990-01-15", // ISO 8601 format (YYYY-MM-DD)
  "city": "Mumbai"
}
```

**Response:**
```json
{
  "message": "Profile updated successfully",
  "user": { ... }
}
```

### 4. Verify PAN (`POST /api/v1/auth/verify-pan`) - **NEW**

**Requires:** Authentication token

**Request:**
```json
{
  "pan": "ABCDE1234F"
}
```

**Response:**
```json
{
  "message": "PAN verified successfully",
  "panDetails": {
    "name": "John Doe",
    "address": "123 Main St, Mumbai",
    "mobile": "1234567890"
  }
}
```

**Note:** Currently returns mock data. Integrate with actual PAN verification service in production.

### 5. Set Password (`POST /api/v1/auth/set-password`) - **NEW**

**Requires:** Authentication token

**Request:**
```json
{
  "password": "SecurePassword123!",
  "confirmPassword": "SecurePassword123!"
}
```

**Response:**
```json
{
  "message": "Password set successfully",
  "token": "jwt-token",
  "user": { ... }
}
```

## Domain Model Changes

### User Entity

Added new fields:
- `DateOfBirth *time.Time`
- `City *string`
- `PAN *string`
- `PANDetails *PANDetails`

New methods:
- `SetPassword(passwordHash string)` - Sets user password
- `UpdateProfile(phone, dateOfBirth, city)` - Updates profile fields
- `SetPAN(pan string, details *PANDetails)` - Sets PAN and verification details

## Database Changes

### New Fields in `users` Table

The following fields have been added to the database via migration `002_add_user_profile_fields`:

- `date_of_birth DATE` - User's date of birth
- `city VARCHAR(100)` - User's city
- `pan VARCHAR(10)` - PAN number (unique)
- `pan_name VARCHAR(255)` - Name from PAN verification
- `pan_address TEXT` - Address from PAN verification
- `pan_mobile VARCHAR(20)` - Mobile from PAN verification

The `password_hash` column has been made nullable to support step-by-step signup.

**Migration Files:**
- `migrations/postgres/002_add_user_profile_fields.up.sql` - Adds the new fields
- `migrations/postgres/002_add_user_profile_fields.down.sql` - Rolls back the changes

**To run the migration:**
```bash
# Using golang-migrate
migrate -path migrations/postgres -database "postgres://user:password@localhost:5432/equitywala?sslmode=disable" up

# Or using psql
psql -U postgres -d equitywala -f migrations/postgres/002_add_user_profile_fields.up.sql
```

## Testing

All existing tests have been updated to reflect the new API structure. The test suite should pass with the new changes.

## Next Steps

1. **Create Database Migration**: Add migration file for new user fields
2. **PAN Verification Integration**: Replace mock PAN verification with actual service
3. **Update Frontend**: Ensure frontend calls match the new API structure
4. **Email Template**: Already updated to show 4-digit OTP
5. **Documentation**: Update API documentation with new endpoints

## Breaking Changes

⚠️ **Important:** The following are breaking changes:

1. OTP code length changed from 6 to 4 digits
2. Signup no longer requires password (optional)
3. Phone field removed from signup request
4. New authentication-required endpoints added

Frontend must be updated to:
- Use 4-digit OTP codes
- Call new endpoints in the correct order
- Handle step-by-step signup flow


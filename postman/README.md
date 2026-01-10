# Postman Collection for Equitywala Backend API

This directory contains Postman collections and environments for testing the Equitywala Backend API.

## Files

- **Equitywala-Backend.postman_collection.json** - Main Postman collection with all API endpoints
- **Equitywala-Backend-Local.postman_environment.json** - Local development environment variables

## Setup Instructions

### 1. Import Collection and Environment

1. Open Postman
2. Click **Import** button
3. Select both JSON files:
   - `Equitywala-Backend.postman_collection.json`
   - `Equitywala-Backend-Local.postman_environment.json`
4. Click **Import**

### 2. Select Environment

1. In Postman, click the environment dropdown (top right)
2. Select **Equitywala Backend - Local**

### 3. Start the Backend Server

Make sure your backend server is running:

```bash
cd ssit-backend
make run
# or
go run ./cmd/server
```

The server should be running on `http://localhost:8080`

### 4. Test the API

1. Start with the **Health Check** endpoint to verify the server is running
2. Use **Authentication > Signup** or **Login** to get an auth token
3. The auth token will be automatically saved to the environment variable `auth_token`
4. All subsequent requests will use this token automatically

## Collection Structure

The collection is organized into the following folders:

### Health
- **Health Check** - Verify server status

### Authentication
- **Signup** - Register new user
- **Login** - Authenticate and get JWT token
- **Send OTP** - Send OTP to phone number
- **Verify OTP** - Verify OTP code
- **Refresh Token** - Refresh access token
- **Logout** - Logout and invalidate session

### Users
- **Get User by ID** - Get user details
- **Get Current User** - Get authenticated user profile
- **Update User** - Update user profile
- **List Users** - List all users (Admin only)

### KYC
- **Submit KYC** - Submit KYC documents
- **Get KYC Status** - Get verification status
- **Get KYC Details** - Get detailed KYC information
- **Update KYC** - Update KYC information

### Subscriptions
- **Get Subscription Packages** - List available packages
- **Get Current Subscription** - Get user's active subscription
- **Subscribe** - Subscribe to a package
- **Validate Voucher** - Validate discount voucher
- **Get Subscription History** - Get subscription history

### Advisory
- **Get Advisory List** - List advisory content
- **Get Advisory by ID** - Get specific advisory
- **Create Advisory** - Create new advisory (Admin/Advisor)
- **Update Advisory** - Update advisory (Admin/Advisor)
- **Delete Advisory** - Delete advisory (Admin)

### Payments
- **Create Payment** - Create payment order
- **Get Payment Status** - Check payment status
- **Get Payment History** - Get payment history
- **Payment Webhook** - Payment gateway webhook

### Admin
- **List All Users** - List all users with filters
- **Update User Role** - Change user role
- **Verify KYC** - Manually verify KYC
- **Get Dashboard Stats** - Get admin statistics

## Environment Variables

The collection uses the following environment variables:

- `base_url` - API base URL (default: `http://localhost:8080`)
- `auth_token` - JWT authentication token (auto-populated after login)
- `refresh_token` - Refresh token for token renewal (auto-populated after login)
- `user_id` - Current user ID (optional)

## Auto Token Management

The collection includes scripts that automatically:
- Save `auth_token` after successful login/signup
- Save `refresh_token` after login
- Use `auth_token` in Authorization header for protected endpoints

## Notes

- Most endpoints require authentication via JWT token
- The token is automatically included in requests that need it
- Some endpoints are role-based (Admin, Advisor, User)
- Request/response examples are included in the collection
- Update the `base_url` variable for different environments (staging, production)

## Creating Additional Environments

To create environments for staging or production:

1. Duplicate the Local environment file
2. Update the `base_url` value
3. Import it into Postman
4. Select the appropriate environment when testing

Example for production:
```json
{
  "key": "base_url",
  "value": "https://api.equitywala.com",
  ...
}
```










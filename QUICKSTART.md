# Quick Start Guide

## What Has Been Created

### ✅ Project Structure
- Complete Clean Architecture directory structure
- Go module setup with dependencies
- Configuration management
- Database connection setup (PostgreSQL + MongoDB)

### ✅ Core Components

1. **Domain Layer** (`internal/domain/`)
   - `user` module: User entity, repository interface, domain errors
   - `auth` module: OTP and Session entities, repository interfaces

2. **Application Layer** (`internal/application/`)
   - Example command handler: `CreateUserCommand`
   - Example query handler: `GetUserQuery`
   - Demonstrates Command/Query pattern

3. **Infrastructure Layer** (`internal/infrastructure/`)
   - PostgreSQL connection with GORM
   - MongoDB connection
   - Database connection pooling configured

4. **Interface Layer** (`internal/interfaces/`)
   - API wrapper (`api.go`, `bind.go`) - Standard response format
   - Route setup with placeholder endpoints
   - Health check endpoint

5. **Shared Utilities** (`internal/shared/`)
   - Error handling (DomainError)
   - Structured logging (Zap)

### ✅ Configuration
- Environment variable support
- Database configuration
- JWT configuration
- Logging configuration

## Next Steps

### 1. Install Dependencies

```bash
cd ssit-backend
go mod download
go mod tidy
```

### 2. Set Up Environment

Create a `.env` file (copy from `.env.example` when available):

```env
SERVER_PORT=8080
SERVER_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=equitywala
MONGODB_URI=mongodb://localhost:27017
JWT_SECRET=your-secret-key-change-in-production
```

### 3. Set Up Databases

**PostgreSQL:**
```bash
# Create database
createdb equitywala

# Or using psql
psql -U postgres -c "CREATE DATABASE equitywala;"
```

**MongoDB:**
```bash
# Start MongoDB (if using local)
mongod
```

### 4. Run the Server

```bash
# Using Make
make run

# Or directly
go run ./cmd/server
```

The server will start on `http://localhost:8080`

### 5. Test Health Endpoint

```bash
curl http://localhost:8080/health
```

## Implementation Roadmap

### Phase 1: Complete Auth Module
- [ ] Implement OTP generation and verification
- [ ] Implement JWT token generation
- [ ] Implement session management
- [ ] Create auth handlers (signup, login, OTP verification)
- [ ] Create auth middleware

### Phase 2: User Repository Implementation
- [ ] Implement PostgreSQL user repository
- [ ] Add GORM models
- [ ] Write repository tests
- [ ] Implement user classification logic

### Phase 3: KYC Module
- [ ] Create KYC domain entities
- [ ] Implement KYC repository
- [ ] Create KYC command/query handlers
- [ ] Integrate with KYC provider

### Phase 4: Subscription Module
- [ ] Create subscription domain entities
- [ ] Implement subscription repository
- [ ] Create pricing package management
- [ ] Implement voucher system

### Phase 5: Payment Module
- [ ] Create payment domain entities
- [ ] Integrate payment gateway
- [ ] Implement webhook handling

### Phase 6: Advisory Module
- [ ] Create advisory domain entities
- [ ] Implement advisory repository
- [ ] Create advisory publishing handlers
- [ ] Implement access control

## Code Patterns to Follow

### Command Handler Pattern

```go
// 1. Define command
type CreateUserCommand struct {
    Email    string
    Password string
}

// 2. Create handler
type CreateUserHandler struct {
    repo user.Repository
}

// 3. Implement Handle method
func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) error {
    // Validation
    // Business logic
    // Persistence
    return nil
}
```

### Query Handler Pattern

```go
// 1. Define query
type GetUserQuery struct {
    UserID uuid.UUID
}

// 2. Define result
type GetUserResult struct {
    User *user.User
}

// 3. Create handler
func (h *GetUserHandler) Handle(ctx context.Context, query GetUserQuery) (*GetUserResult, error) {
    // Fetch data
    return &GetUserResult{User: user}, nil
}
```

### API Handler Pattern

```go
// Handler returns (interface{}, error)
func (h *UserHandler) CreateUser(ctx *api.Context) (interface{}, error) {
    var req dto.CreateUserRequest
    
    // Bind and validate
    if err := ctx.BindJSON(&req); err != nil {
        return nil, err
    }
    
    // Execute command
    cmd := commands.CreateUserCommand{...}
    if err := h.createUserHandler.Handle(ctx.Request.Context(), cmd); err != nil {
        return nil, err
    }
    
    // Return response
    return dto.CreateUserResponse{...}, nil
}
```

## Testing

### Run Tests
```bash
make test
```

### Run with Coverage
```bash
make test-coverage
```

## Development Guidelines

1. **Follow TDD**: Write tests first
2. **Use Command/Query Pattern**: Commands for writes, Queries for reads
3. **Domain-Driven Design**: Keep business logic in domain layer
4. **Error Handling**: Use domain errors from `shared/errors`
5. **Logging**: Use structured logging from `shared/logger`
6. **API Responses**: Always use `api.Handle` wrapper

## Resources

- [Backend Rules](../backend-rules.md) - Complete development guidelines
- [Database Schema Plan](../knowledge-based/database-schema-plan.md) - Database structure
- [API Documentation](./docs/API.md) - API endpoint documentation

---

**Happy Coding! 🚀**


# Equitywala Stock Advisory Platform - Backend

**Version:** 1.0.0  
**Tech Stack:** Golang, Gin, GORM, PostgreSQL, MongoDB

## Project Structure

```
ssit-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/                     # Configuration management
│   ├── domain/                     # Domain entities and interfaces
│   │   ├── auth/
│   │   ├── kyc/
│   │   ├── subscription/
│   │   ├── advisory/
│   │   ├── payment/
│   │   ├── user/
│   │   └── audit/
│   ├── application/                # Application services (use cases)
│   │   ├── commands/              # Command handlers
│   │   ├── queries/               # Query handlers
│   │   └── services/              # Application services
│   ├── infrastructure/            # External concerns
│   │   ├── database/
│   │   │   ├── postgres/          # PostgreSQL repositories
│   │   │   └── mongodb/           # MongoDB repositories
│   │   ├── messaging/             # Message queues, events
│   │   ├── external/              # External API clients (KYC, Payment)
│   │   └── cache/                 # Caching layer
│   ├── interfaces/                # HTTP handlers, middleware
│   │   ├── http/
│   │   │   ├── api/               # API wrapper (api.go, bind.go)
│   │   │   ├── handlers/          # HTTP handlers
│   │   │   ├── middleware/        # HTTP middleware
│   │   │   ├── routes/            # Route definitions
│   │   │   └── dto/               # Data Transfer Objects
│   │   └── cli/                   # CLI commands (if any)
│   └── shared/                     # Shared utilities
│       ├── errors/
│       ├── logger/
│       ├── validator/
│       └── utils/
├── pkg/                            # Public packages (reusable)
│   ├── errors/
│   └── utils/
├── migrations/                     # Database migrations
│   ├── postgres/
│   └── mongodb/
├── tests/                          # Integration and E2E tests
│   ├── integration/
│   └── e2e/
├── scripts/                        # Build and deployment scripts
├── docs/                           # API documentation
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+ (for local development)
- MongoDB 6+ (for audit logs)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run database migrations manually (see `migrations/README.md`):
   ```bash
   # Use migration tools or run SQL files manually
   ```

5. Start the server:
   ```bash
   make run
   ```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Code Generation

```bash
# Generate mocks
make generate-mocks
```

## Architecture

This project follows **Clean Architecture** principles with:

- **Domain Layer**: Core business entities and interfaces
- **Application Layer**: Use cases (commands/queries)
- **Infrastructure Layer**: External implementations (database, APIs)
- **Interface Layer**: HTTP handlers and DTOs

See `backend-rules.md` for detailed development guidelines.

## API Documentation

API documentation will be available at `/api/docs` once Swagger is integrated.

## License

Proprietary - Equitywala


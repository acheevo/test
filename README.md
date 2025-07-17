# Test API

A basic Go REST API built following the Nimbus architecture patterns.

## Features

- User authentication (JWT-based sessions)
- User management
- Clean architecture with repository pattern
- Middleware for authentication
- CORS support
- PostgreSQL database with GORM

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `POST /api/auth/logout` - User logout

### Users (Protected)
- `GET /api/users/me` - Get current user
- `GET /api/users` - Get all users (admin only)
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create new user (admin only)

### Health Check
- `GET /health` - Health check endpoint

## Getting Started

1. Install dependencies:
   ```bash
   make deps
   ```

2. Set up environment variables:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=testuser
   export DB_PASSWORD=testpass
   export DB_NAME=testdb
   export ADMIN_EMAIL=admin@test.local
   export ADMIN_PASSWORD=admin123
   ```

3. Run the application:
   ```bash
   make run
   ```

4. Or run with Docker:
   ```bash
   make docker-run
   ```

## Testing

- Run all tests: `make test`
- Run unit tests: `make test-unit`
- Run integration tests: `make test-integration`
- Run with coverage: `make test-coverage`

## Docker

- Build image: `make docker-build`
- Run with compose: `make docker-run`
- Stop containers: `make docker-stop`
- Clean up: `make docker-clean`

## Architecture

The project follows a clean architecture pattern with:

- `cmd/api/` - Application entry point
- `internal/config/` - Configuration management
- `internal/models/` - Data models
- `internal/repository/` - Database layer
- `internal/services/` - Business logic
- `internal/handlers/` - HTTP handlers
- `internal/middleware/` - HTTP middleware
- `internal/http/` - HTTP server setup

## Database Schema

The API uses PostgreSQL with the following tables:
- `users` - User accounts with roles
- `sessions` - Authentication sessions

# Auth Service

A comprehensive authentication service built with Go, featuring JWT-based authentication, PostgreSQL for user storage, and Redis for session/token caching.

## Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: Bcrypt for secure password storage
- **PostgreSQL**: User data storage
- **Redis**: Session and token caching
- **RESTful API**: Clean and intuitive endpoints
- **CORS Support**: Cross-origin resource sharing enabled

## Endpoints

### Public Endpoints (No Authentication Required)

- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token

### Protected Endpoints (Authentication Required)

- `POST /auth/logout` - User logout
- `GET /auth/profile` - Get user profile

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository and navigate to the auth-service directory
2. Run the services:
   ```bash
   docker-compose up -d
   ```
3. The service will be available at `http://localhost:8080`

### Manual Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up environment variables (create a `.env` file):
   ```env
   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=password
   DB_NAME=auth_service

   # Redis Configuration
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0

   # JWT Configuration
   JWT_SECRET=your-secret-key-change-in-production

   # Server Configuration
   PORT=8080
   ```

3. Start PostgreSQL and Redis
4. Run the service:
   ```bash
   go run main.go
   ```

## API Documentation

### Register User

**POST** `/auth/register`

Request body:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

Response:
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "jwt_token",
  "refresh_token": "jwt_refresh_token",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### Login

**POST** `/auth/login`

Request body:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Response: Same as register response

### Refresh Token

**POST** `/auth/refresh`

Request body:
```json
{
  "refresh_token": "jwt_refresh_token"
}
```

Response: Same as register response with new tokens

### Logout

**POST** `/auth/logout`

Headers:
```
Authorization: Bearer <access_token>
```

Response:
```json
{
  "message": "Logged out successfully"
}
```

### Get Profile

**GET** `/auth/profile`

Headers:
```
Authorization: Bearer <access_token>
```

Response:
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

## Health Check

**GET** `/health`

Response:
```json
{
  "status": "ok",
  "service": "auth-service"
}
```

## Security Features

- **Password Hashing**: Passwords are hashed using bcrypt with default cost
- **JWT Tokens**: Access tokens expire in 15 minutes, refresh tokens in 7 days
- **Token Storage**: Refresh tokens are stored in Redis for additional security
- **Input Validation**: All inputs are validated using Gin's binding
- **CORS**: Configured for cross-origin requests

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | PostgreSQL user | postgres |
| `DB_PASSWORD` | PostgreSQL password | password |
| `DB_NAME` | PostgreSQL database name | auth_service |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_DB` | Redis database number | 0 |
| `JWT_SECRET` | JWT signing secret | your-secret-key-change-in-production |
| `PORT` | Server port | 8080 |

## Development

### Prerequisites

- Go 1.24.5 or higher
- PostgreSQL 15 or higher
- Redis 7 or higher

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o auth-service main.go
```

## License

MIT License 
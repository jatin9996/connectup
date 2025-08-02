# Auth Service with Matchmaker

A comprehensive authentication service built with Go, featuring JWT-based authentication, PostgreSQL for user storage, Redis for session/token caching, and an integrated matchmaker service for user matching based on tags, industries, experience, and interests.

## Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: Bcrypt for secure password storage
- **PostgreSQL**: User data storage
- **Redis**: Session and token caching
- **RESTful API**: Clean and intuitive endpoints
- **CORS Support**: Cross-origin resource sharing enabled
- **Matchmaker Service**: Intelligent user matching based on tags, industries, experience, and interests
- **Kafka Integration**: Event-driven architecture for user updates and match creation
- **Real-time Matching**: Automatic match generation when user profiles are updated

## Endpoints

### Public Endpoints (No Authentication Required)

- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token

### Protected Endpoints (Authentication Required)

- `POST /auth/logout` - User logout
- `GET /auth/profile` - Get user profile

### Matchmaker Endpoints

- `POST /api/v1/matchmaker/profiles` - Create user profile for matchmaking
- `GET /api/v1/matchmaker/profiles/:user_id` - Get user profile
- `GET /api/v1/matchmaker/matches/:user_id` - Get matches for a user
- `GET /api/v1/matchmaker/matches/details/:match_id` - Get match details
- `PUT /api/v1/matchmaker/matches/:match_id/status` - Update match status
- `POST /api/v1/matchmaker/search` - Search for matches based on criteria

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

   # Kafka Configuration
   KAFKA_BROKERS=localhost:9092
   KAFKA_USER_UPDATED_TOPIC=user-updated
   ```

3. Start PostgreSQL, Redis, and Kafka
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

## Matchmaker API Documentation

### Create User Profile

**POST** `/api/v1/matchmaker/profiles`

Request body:
```json
{
  "user_id": "user123",
  "tags": ["golang", "backend", "microservices"],
  "industries": ["technology", "software", "fintech"],
  "experience": 5,
  "interests": ["open source", "cloud computing"],
  "location": "San Francisco, CA",
  "bio": "Backend developer with 5 years of experience",
  "skills": ["Go", "PostgreSQL", "Redis", "Docker"]
}
```

Response:
```json
{
  "message": "User profile created successfully",
  "matches_found": 3
}
```

### Get User Profile

**GET** `/api/v1/matchmaker/profiles/:user_id`

Response:
```json
{
  "profile": {
    "user_id": "user123",
    "tags": ["golang", "backend", "microservices"],
    "industries": ["technology", "software", "fintech"],
    "experience": 5,
    "interests": ["open source", "cloud computing"],
    "location": "San Francisco, CA",
    "bio": "Backend developer with 5 years of experience",
    "skills": ["Go", "PostgreSQL", "Redis", "Docker"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Get Matches

**GET** `/api/v1/matchmaker/matches/:user_id?status=pending&limit=10&offset=0`

Response:
```json
{
  "matches": [
    {
      "id": "match123",
      "user_id_1": "user123",
      "user_id_2": "user456",
      "score": 0.85,
      "common_tags": ["golang", "backend"],
      "common_skills": ["Go", "PostgreSQL"],
      "status": "pending",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

### Update Match Status

**PUT** `/api/v1/matchmaker/matches/:match_id/status`

Request body:
```json
{
  "status": "accepted"
}
```

Response:
```json
{
  "message": "Match status updated successfully",
  "match": {
    "id": "match123",
    "user_id_1": "user123",
    "user_id_2": "user456",
    "score": 0.85,
    "common_tags": ["golang", "backend"],
    "common_skills": ["Go", "PostgreSQL"],
    "status": "accepted",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Search Matches

**POST** `/api/v1/matchmaker/search`

Request body:
```json
{
  "user_id": "user123",
  "industries": ["technology", "software"],
  "min_exp": 3,
  "max_exp": 7,
  "skills": ["Go", "PostgreSQL"],
  "location": "San Francisco",
  "limit": 10,
  "offset": 0
}
```

Response:
```json
{
  "matches": [
    {
      "user_id": "user456",
      "score": 0.85,
      "reason": "Common interests: golang, backend; Common skills: Go, PostgreSQL; Similar experience level"
    }
  ],
  "total": 1
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
| `KAFKA_BROKERS` | Kafka broker addresses | localhost:9092 |
| `KAFKA_USER_UPDATED_TOPIC` | Kafka topic for user updates | user-updated |

## Development

### Prerequisites

- Go 1.24.5 or higher
- PostgreSQL 15 or higher
- Redis 7 or higher
- Kafka 7.4.0 or higher

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o auth-service main.go
```

### Testing the Matchmaker

1. Start the services:
   ```bash
   docker-compose up -d
   ```

2. Run the test script to publish sample user events:
   ```bash
   go run examples/matchmaker_test.go
   ```

3. Test the REST endpoints:
   ```bash
   # Get matches for a user
   curl http://localhost:8080/api/v1/matchmaker/matches/user1

   # Get user profile
   curl http://localhost:8080/api/v1/matchmaker/profiles/user1

   # Search for matches
   curl -X POST http://localhost:8080/api/v1/matchmaker/search \
     -H "Content-Type: application/json" \
     -d '{"user_id": "user1", "limit": 10, "offset": 0}'
   ```

## Matchmaker Algorithm

The matchmaker service uses a weighted scoring algorithm based on:

- **Tags Similarity (30%)**: Jaccard similarity of user tags
- **Industry Similarity (25%)**: Jaccard similarity of industries
- **Experience Compatibility (20%)**: Experience level compatibility
- **Skills Similarity (15%)**: Jaccard similarity of skills
- **Location Compatibility (10%)**: Geographic proximity

Matches are triggered automatically when user profiles are updated via Kafka events.

## License

MIT License 
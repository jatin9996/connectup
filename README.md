# Auth Service - Enhanced with Showcase & Real-time Messaging

A comprehensive authentication and showcase service built with Go, featuring real-time messaging, company profile management, and investment tracking.

## 🚀 Features

### Core Authentication
- User registration and login
- JWT token-based authentication
- Password hashing and validation
- Session management

### Showcase Service
- **Company Profile Management**: Create, update, and manage company profiles
- **Investment Tracking**: Record and track investments with detailed metrics
- **Admin/Investor APIs**: Secure endpoints for authorized users only
- **Public Company Discovery**: Search and browse public company profiles
- **Analytics Integration**: Track user interactions and company views

### Real-time Messaging
- **WebSocket Support**: Real-time chat functionality
- **Kafka Integration**: Message queuing and event streaming
- **Typing Indicators**: Real-time typing status
- **Read Receipts**: Message delivery confirmation
- **Online Status**: User presence tracking

### Data Management
- **PostgreSQL Database**: Robust data storage with optimized indexes
- **Redis Caching**: Cache popular company profiles for performance
- **Full-text Search**: Advanced search capabilities with GIN indexes
- **Data Analytics**: Comprehensive event tracking and analytics

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   Admin Panel   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Auth Service  │
                    │   (Go/Gin)      │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │     Redis       │    │     Kafka       │
│   (Primary DB)  │    │   (Caching)     │    │  (Event Bus)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📊 Database Schema

### Core Tables
- `users` - User accounts and authentication
- `companies` - Company profiles and information
- `investments` - Investment records and metrics
- `messages` - Chat messages and conversations
- `analytics_events` - User interaction tracking
- `sessions` - WebSocket session management

### Key Features
- **UUID Primary Keys**: Secure and globally unique identifiers
- **Full-text Search**: GIN indexes for company name and description
- **Foreign Key Constraints**: Data integrity and referential integrity
- **JSONB Support**: Flexible analytics data storage
- **Optimized Indexes**: Performance-optimized queries

## 🔧 Setup & Installation

### Prerequisites
- Go 1.24.5+
- PostgreSQL 12+
- Redis 6+
- Kafka 2.8+

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=auth_service

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_USER_UPDATED_TOPIC=user-updated
KAFKA_CHAT_TOPIC=chat-messages
KAFKA_ANALYTICS_TOPIC=analytics_events

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Server
PORT=8080
```

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd auth-service

# Install dependencies
go mod tidy

# Run the service
go run main.go
```

## 📡 API Endpoints

### Authentication
```
POST   /api/v1/auth/register     # User registration
POST   /api/v1/auth/login        # User login
POST   /api/v1/auth/logout       # User logout
GET    /api/v1/auth/profile      # Get user profile
PUT    /api/v1/auth/profile      # Update user profile
```

### Showcase Service (Authenticated)
```
POST   /api/v1/showcase/companies           # Create company profile
GET    /api/v1/showcase/companies/:id       # Get company profile
PUT    /api/v1/showcase/companies/:id       # Update company profile
GET    /api/v1/showcase/companies           # Search companies

POST   /api/v1/showcase/investments         # Create investment record
GET    /api/v1/showcase/companies/:id/investments  # Get company investments
GET    /api/v1/showcase/investments/my      # Get user investments

POST   /api/v1/showcase/analytics/events    # Track analytics events
```

### Showcase Service (Public)
```
GET    /api/v1/showcase/public/companies    # Search public companies
GET    /api/v1/showcase/public/companies/:id # Get public company profile
```

### WebSocket
```
GET    /ws                    # WebSocket connection
GET    /api/v1/websocket/online-users  # Get online users
```

### Matchmaker Service
```
POST   /api/v1/matchmaker/profiles          # Create user profile
GET    /api/v1/matchmaker/profiles/:user_id # Get user profile
GET    /api/v1/matchmaker/matches/:user_id  # Get user matches
PUT    /api/v1/matchmaker/matches/:match_id/status # Update match status
POST   /api/v1/matchmaker/search            # Search matches
```

## 💬 WebSocket Messaging

### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onopen = () => {
    console.log('Connected to WebSocket');
};
```

### Message Types
```javascript
// Send chat message
ws.send(JSON.stringify({
    type: 'chat_message',
    receiver_id: 'user-uuid',
    content: 'Hello!'
}));

// Typing indicator
ws.send(JSON.stringify({
    type: 'typing',
    receiver_id: 'user-uuid',
    is_typing: true
}));

// Read receipt
ws.send(JSON.stringify({
    type: 'read_receipt',
    message_id: 'message-uuid'
}));
```

### Message Events
```javascript
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    switch(data.type) {
        case 'connection_established':
            console.log('Connection established');
            break;
        case 'chat_message':
            console.log('New message:', data.message);
            break;
        case 'typing_indicator':
            console.log('User typing:', data.user_id);
            break;
        case 'read_receipt':
            console.log('Message read:', data.message_id);
            break;
    }
};
```

## 🏢 Company Profile Management

### Create Company Profile
```bash
curl -X POST http://localhost:8080/api/v1/showcase/companies \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TechCorp Inc",
    "description": "Innovative technology company",
    "industry": "Technology",
    "founded_year": 2020,
    "headquarters": "San Francisco, CA",
    "website": "https://techcorp.com",
    "employee_count": 150,
    "revenue": 5000000,
    "funding_stage": "Series A",
    "total_funding": 2000000,
    "valuation": 25000000,
    "is_public": true
  }'
```

### Search Companies
```bash
curl "http://localhost:8080/api/v1/showcase/companies?q=tech&industry=Technology&limit=10&offset=0"
```

## 💰 Investment Tracking

### Create Investment
```bash
curl -X POST http://localhost:8080/api/v1/showcase/investments \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "company_id": "company-uuid",
    "amount": 500000,
    "currency": "USD",
    "investment_type": "equity",
    "round": "Series A",
    "date": "2024-01-15",
    "status": "completed",
    "notes": "Strategic investment"
  }'
```

## 📈 Analytics & Events

### Track Custom Events
```bash
curl -X POST http://localhost:8080/api/v1/showcase/analytics/events \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "company_viewed",
    "company_id": "company-uuid",
    "view_duration": 45,
    "source": "search_results"
  }'
```

## 🔒 Security Features

### Authentication
- JWT-based authentication with configurable expiry
- Password hashing using bcrypt
- Session management with Redis
- Role-based access control (admin/investor)

### Data Protection
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- CORS configuration
- Rate limiting (can be added)

## 🚀 Performance Optimizations

### Caching Strategy
- **Redis Caching**: Popular company profiles cached for 1 hour
- **Database Indexes**: Optimized queries with strategic indexing
- **Connection Pooling**: Efficient database connection management

### Scalability Features
- **Kafka Integration**: Event-driven architecture for scalability
- **WebSocket Connections**: Efficient real-time messaging
- **Stateless Design**: Horizontal scaling capability
- **Microservice Ready**: Modular architecture for service decomposition

## 🧪 Testing

### Run Tests
```bash
go test ./...
```

### Test Coverage
```bash
go test -cover ./...
```

## 📦 Docker Deployment

### Docker Compose
```yaml
version: '3.8'
services:
  auth-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - postgres
      - redis
      - kafka

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: auth_service
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:7.4.0
    environment:
      KAFKA_CFG_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_CFG_LISTENERS: PLAINTEXT://0.0.0.0:9092
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181

volumes:
  postgres_data:
```

## 🔧 Configuration

### Production Settings
- Set appropriate JWT secrets
- Configure database connection pooling
- Enable SSL/TLS for WebSocket connections
- Set up proper CORS origins
- Configure Kafka topics and partitions
- Set up monitoring and logging

### Monitoring
- Health check endpoint: `GET /health`
- Service metrics and logging
- Database connection monitoring
- Kafka consumer lag monitoring
- Redis memory usage monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the documentation
- Review the API examples

---

**Built with ❤️ using Go, Gin, PostgreSQL, Redis, and Kafka** 
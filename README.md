# Hybrid REST + gRPC User Management Service

A production-ready microservice that provides both REST and gRPC APIs for user management, with shared business logic, comprehensive monitoring, and health checks.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Hybrid Service                           │
├─────────────────────────────────────────────────────────────┤
│  REST Server (:8080)    │    gRPC Server (:50051)         │
│  ├─ /signup            │    ├─ CreateUser                 │
│  ├─ /login             │    ├─ GetUser                    │
│  ├─ /users/*           │    ├─ UpdateUser                 │
│  ├─ /healthz           │    ├─ DeleteUser                 │
│  └─ /metrics           │    └─ ListUsers                  │
├─────────────────────────────────────────────────────────────┤
│                Shared Business Logic                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              UserService                            │   │
│  │  ├─ CreateUser()                                   │   │
│  │  ├─ GetUser()                                      │   │
│  │  ├─ UpdateUser()                                   │   │
│  │  ├─ DeleteUser()                                   │   │
│  │  └─ ListUsers()                                    │   │
│  └─────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                Database Layer                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  PostgreSQL + GORM + Retry Logic + Metrics         │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
restapi/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go          # REST API handlers
│   │   └── middleware.go        # HTTP middleware
│   ├── grpc/
│   │   └── grpc_server.go       # gRPC server implementation
│   ├── service/
│   │   └── service.go           # Business logic layer
│   ├── database/
│   │   └── database.go          # Database operations
│   ├── metrics/
│   │   └── metrics.go           # Prometheus metrics
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   └── retry/
│       └── retry.go             # Retry logic
├── pkg/
│   ├── models/
│   │   └── user.go              # Data models
│   └── proto/
│       ├── user.proto           # gRPC service definition
│       ├── user.pb.go           # Generated protobuf messages
│       └── user_grpc.pb.go      # Generated gRPC service code
├── scripts/
│   ├── test_complete.sh         # Complete test script
│   ├── test_simple.sh           # Simple test script
│   └── test_grpc_simple.sh      # gRPC test script
├── docs/
│   ├── README_HYBRID.md         # Detailed documentation
│   ├── TESTING_GUIDE.md         # Testing guide
│   └── IMPLEMENTATION_SUMMARY.md # Implementation summary
├── Makefile                     # Build automation
├── Dockerfile                   # Container definition
├── docker-compose.yml           # Multi-service setup
├── prometheus.yml               # Metrics configuration
└── go.mod                       # Go module dependencies
```

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL database
- protoc (Protocol Buffers compiler)

### Installation

1. **Clone and setup**:
```bash
git clone <repository>
cd restapi
make deps
make dev-tools
```

2. **Start PostgreSQL**:
```bash
# Using Docker
docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=restapi -p 5432:5432 -d postgres:15
```

3. **Run the service**:
```bash
make run
```

4. **Test the service**:
```bash
make test-script
```

### Using Docker Compose
```bash
make docker-run
```

## 📊 API Endpoints

### REST API (Port 8080)

#### Public Endpoints
- `POST /signup` - User registration
- `POST /login` - User authentication

#### Protected Endpoints (Require JWT)
- `GET /users` - List all users
- `GET /users/:id` - Get user by ID
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user

#### System Endpoints
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics

### gRPC API (Port 50051)

#### Service: `user.UserService`
- `CreateUser(CreateUserRequest) → UserResponse`
- `GetUser(GetUserRequest) → UserResponse`
- `UpdateUser(UpdateUserRequest) → UserResponse`
- `DeleteUser(DeleteUserRequest) → DeleteUserResponse`
- `ListUsers(ListUsersRequest) → ListUsersResponse`

## 🔧 Development

### Available Make Targets

```bash
make help          # Show available targets
make build         # Build the application
make run           # Run the application
make test          # Run tests
make clean         # Clean build artifacts
make proto         # Generate protobuf code
make docker-build  # Build Docker image
make docker-run    # Run with Docker Compose
make deps          # Install dependencies
make dev-tools     # Install development tools
make test-script   # Run test script
```

### Adding New Features

1. **Add new gRPC method**:
   - Update `pkg/proto/user.proto`
   - Run `make proto`
   - Implement in `internal/grpc/grpc_server.go`

2. **Add new REST endpoint**:
   - Add route in `cmd/server/main.go`
   - Implement handler in `internal/api/handlers.go`

3. **Add new business logic**:
   - Add method to `internal/service/service.go`
   - Use in both REST and gRPC handlers

## 📈 Monitoring

### Prometheus Metrics

- **HTTP Metrics**: `http_requests_total`, `http_request_duration_seconds`
- **gRPC Metrics**: `grpc_requests_total`, `grpc_request_duration_seconds`
- **Database Metrics**: `db_operations_total`, `db_operation_duration_seconds`
- **Health Metrics**: `health_check_status`

### Health Checks
- **Endpoint**: `GET /healthz`
- **Checks**: Database connectivity
- **Response**: JSON with status, timestamp, and service health

## 🐳 Docker

### Build and Run
```bash
# Build image
make docker-build

# Run with Docker Compose (includes PostgreSQL and Prometheus)
make docker-run

# Stop services
make docker-stop
```

### Environment Variables
- `DATABASE_URL` - PostgreSQL connection string
- `ENV` - Environment (production/development)

## 🧪 Testing

### Manual Testing
```bash
# Run complete test suite
make test-script

# Test specific components
./scripts/test_simple.sh      # Basic REST API test
./scripts/test_grpc_simple.sh # gRPC connectivity test
```

### Using grpcurl
```bash
# List services
grpcurl -plaintext localhost:50051 list

# Create user
grpcurl -plaintext -d '{"name":"Test User","email":"test@example.com","password":"password123"}' \
  localhost:50051 user.UserService/CreateUser

# List users
grpcurl -plaintext localhost:50051 user.UserService/ListUsers
```

## 🔒 Security

- JWT-based authentication for REST API
- Password hashing with bcrypt
- Input validation and sanitization
- Structured logging for audit trails

## 📚 Documentation

- [Detailed Implementation Guide](docs/README_HYBRID.md)
- [Testing Guide](docs/TESTING_GUIDE.md)
- [Implementation Summary](docs/IMPLEMENTATION_SUMMARY.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

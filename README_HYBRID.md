# Hybrid REST + gRPC User Management Service

A hybrid microservice that provides both REST and gRPC APIs for user management, with shared business logic, comprehensive monitoring, and health checks.

## Features

### v1 Implementation
- **Dual API Support**: Both REST (Gin) and gRPC servers running in the same binary
- **Shared Business Logic**: Common UserService used by both REST and gRPC handlers
- **Comprehensive Monitoring**: Prometheus metrics for both REST and gRPC requests
- **Health Checks**: Database connectivity monitoring via `/healthz`
- **Retry Logic**: Built-in retry mechanisms for database operations
- **Structured Logging**: Comprehensive logging with context and correlation IDs

## Architecture

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

## API Endpoints

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

## Monitoring & Metrics

### Prometheus Metrics

#### HTTP Metrics
- `http_requests_total` - Total HTTP requests by method, endpoint, status
- `http_request_duration_seconds` - HTTP request duration histogram

#### gRPC Metrics
- `grpc_requests_total` - Total gRPC requests by method, status
- `grpc_request_duration_seconds` - gRPC request duration histogram

#### Database Metrics
- `db_operations_total` - Database operations by operation, table, status
- `db_operation_duration_seconds` - Database operation duration histogram

#### Health Metrics
- `health_check_status` - Health check status (1=healthy, 0=unhealthy)

### Health Checks
- **Endpoint**: `GET /healthz`
- **Checks**: Database connectivity
- **Response**: JSON with status, timestamp, and service health

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL database
- protoc (Protocol Buffers compiler)

### Installation

1. **Clone and build**:
```bash
cd packages/restapi
go mod tidy
go build -o hybrid-api .
```

2. **Start PostgreSQL**:
```bash
# Using Docker
docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=restapi -p 5432:5432 -d postgres:15
```

3. **Run the service**:
```bash
./hybrid-api
```

4. **Test the service**:
```bash
./test_hybrid.sh
```

### Configuration

Environment variables:
- `DATABASE_URL` - PostgreSQL connection string (default: localhost)
- `ENV` - Environment (production/development, affects logging)

## Usage Examples

### REST API Examples

#### Signup
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'
```

#### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'
```

#### Get Users (with auth)
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8080/users
```

### gRPC Examples

#### Using grpcurl
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Create user
grpcurl -plaintext -d '{"name":"Jane Doe","email":"jane@example.com","password":"password123"}' \
  localhost:50051 user.UserService/CreateUser

# List users
grpcurl -plaintext localhost:50051 user.UserService/ListUsers
```

#### Using Go client
```go
conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
client := NewUserServiceClient(conn)

// Create user
resp, err := client.CreateUser(context.Background(), &CreateUserRequest{
    Name:     "Test User",
    Email:    "test@example.com",
    Password: "password123",
})
```

## Development

### Project Structure
```
restapi/
├── main.go              # Main application entry point
├── grpc_server.go       # gRPC server implementation
├── service.go           # Shared business logic
├── database.go          # Database operations with retry
├── metrics.go           # Prometheus metrics and health checks
├── logger.go            # Structured logging
├── retry.go             # Retry logic implementation
├── user.proto           # gRPC service definition
├── user.pb.go           # Generated protobuf messages
├── user_grpc.pb.go      # Generated gRPC service code
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── test_hybrid.sh       # Test script
└── README_HYBRID.md     # This file
```

### Adding New Features

1. **Add new gRPC method**:
   - Update `user.proto`
   - Regenerate: `protoc --go_out=. --go-grpc_out=. user.proto`
   - Implement in `grpc_server.go`

2. **Add new REST endpoint**:
   - Add route in `main.go`
   - Implement handler function

3. **Add new business logic**:
   - Add method to `UserService` in `service.go`
   - Use in both REST and gRPC handlers

### Monitoring

Access metrics at `http://localhost:8080/metrics` for Prometheus scraping.

Health check at `http://localhost:8080/healthz` for load balancer health checks.

## Production Considerations

1. **Security**:
   - Use proper JWT secrets
   - Enable TLS for gRPC
   - Implement rate limiting

2. **Performance**:
   - Configure connection pooling
   - Add caching layer
   - Optimize database queries

3. **Observability**:
   - Set up Prometheus + Grafana
   - Configure log aggregation
   - Add distributed tracing

4. **Deployment**:
   - Use container orchestration
   - Implement graceful shutdown
   - Add readiness/liveness probes

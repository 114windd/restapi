# Learning Notes: Hybrid REST + gRPC Microservice

## 📚 Project Overview

This project demonstrates building a production-ready microservice that provides both REST and gRPC APIs for user management, sharing the same business logic while maintaining clean architecture principles.

## 🏗️ Architecture Patterns

### 1. Clean Architecture / Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  REST API (Gin)    │    gRPC API (Protocol Buffers)       │
├─────────────────────────────────────────────────────────────┤
│                    Business Logic Layer                     │
│              Service Layer (UserService)                   │
├─────────────────────────────────────────────────────────────┤
│                    Data Access Layer                       │
│              Database Layer (GORM + Retry)                 │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                     │
│  Logging | Metrics | Retry | Database | Proto              │
└─────────────────────────────────────────────────────────────┘
```

**Key Learning**: Each layer has a single responsibility and depends only on inner layers, making the code testable and maintainable.

### 2. Package Organization (Go Best Practices)

```
restapi/
├── cmd/server/          # Application entry points
├── internal/            # Private application code
│   ├── api/            # HTTP handlers and middleware
│   ├── grpc/           # gRPC server implementation
│   ├── service/        # Business logic
│   ├── database/       # Data access layer
│   ├── metrics/        # Monitoring and observability
│   ├── logger/         # Structured logging
│   └── retry/          # Retry mechanisms
├── pkg/                # Public packages
│   ├── models/         # Data models
│   └── proto/          # Protocol buffer definitions
├── scripts/            # Build and test scripts
└── docs/               # Documentation
```

**Key Learning**: 
- `cmd/` for main applications
- `internal/` for private code that can't be imported by other projects
- `pkg/` for public packages that can be imported
- Clear separation of concerns

## 🔧 Key Technologies & Concepts

### 1. Protocol Buffers (gRPC)

**What it is**: A language-neutral, platform-neutral serialization format for structured data.

**Key Files**:
- `pkg/proto/user.proto` - Service definition
- `pkg/proto/user.pb.go` - Generated Go structs
- `pkg/proto/user_grpc.pb.go` - Generated gRPC service code

**Learning Points**:
```protobuf
syntax = "proto3";
package user;
option go_package = "github.com/114windd/restapi/pkg/proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  // ...
}
```

- Define services and messages in `.proto` files
- Generate Go code with `protoc`
- Use `option go_package` to specify Go module path
- gRPC provides type safety and better performance than REST

### 2. Gin Web Framework

**What it is**: A high-performance HTTP web framework for Go.

**Key Patterns**:
```go
// Middleware pattern
r.Use(api.LoggingMiddleware())
r.Use(metrics.PrometheusMiddleware())

// Route grouping
protected := r.Group("/")
protected.Use(api.AuthMiddleware())
{
    protected.GET("/users", api.GetUsers)
}
```

**Learning Points**:
- Middleware for cross-cutting concerns (logging, metrics, auth)
- Route grouping for organization
- Context-based request handling
- JSON binding and validation

### 3. GORM (Object-Relational Mapping)

**What it is**: A fantastic ORM library for Go.

**Key Patterns**:
```go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"not null"`
    Email     string    `json:"email" gorm:"uniqueIndex;not null"`
    Password  string    `json:"-" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Auto-migration
db.AutoMigrate(&models.User{})
```

**Learning Points**:
- Struct tags for database constraints
- Automatic migration
- Query building with method chaining
- Error handling with GORM-specific errors

### 4. JWT Authentication

**What it is**: JSON Web Tokens for stateless authentication.

**Key Patterns**:
```go
func generateJWT(userID uint) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

**Learning Points**:
- Stateless authentication
- Claims-based authorization
- Token expiration handling
- Middleware for protecting routes

### 5. Retry Pattern with Exponential Backoff

**What it is**: A resilience pattern for handling transient failures.

**Key Implementation**:
```go
func ExecuteWithRetry(operation string, fn RetryableFunc, config RetryConfig) error {
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if attempt == config.MaxAttempts {
            return err
        }
        
        delay := calculateDelay(attempt, config.BaseDelay, config.MaxDelay)
        time.Sleep(delay)
    }
}
```

**Learning Points**:
- Exponential backoff prevents overwhelming failing services
- Configurable retry attempts and delays
- Don't retry on business logic errors (e.g., "not found")
- Logging for observability

### 6. Prometheus Metrics

**What it is**: A monitoring and alerting toolkit.

**Key Patterns**:
```go
var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "endpoint", "status_code"},
)

// Usage
httpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
```

**Learning Points**:
- Counter: monotonically increasing values (request counts)
- Histogram: distribution of values (request duration)
- Labels for dimensional data
- Middleware pattern for automatic collection

### 7. Structured Logging

**What it is**: Logging with structured data for better observability.

**Key Patterns**:
```go
logger.Log.WithFields(logrus.Fields{
    "method":  method,
    "path":    path,
    "user_id": userID,
    "type":    "request",
}).Info("Request completed successfully")
```

**Learning Points**:
- Consistent log format across the application
- Contextual information in logs
- Different log levels (Debug, Info, Warn, Error)
- JSON formatting for production

## 🎯 Design Patterns Used

### 1. Service Layer Pattern

**Purpose**: Encapsulate business logic and provide a clean interface.

```go
type UserService struct{}

func (s *UserService) CreateUser(name, email, password string) (*models.User, error) {
    // Business logic here
    // Validation, password hashing, etc.
    return database.CreateUserWithRetry(&user)
}
```

**Benefits**:
- Single responsibility
- Testable business logic
- Reusable across different interfaces (REST, gRPC)

### 2. Repository Pattern (via GORM)

**Purpose**: Abstract data access logic.

```go
func CreateUserWithRetry(user *models.User) error {
    return retry.ExecuteWithRetry("create_user", func() error {
        return db.Create(user).Error
    }, config)
}
```

**Benefits**:
- Database abstraction
- Retry logic encapsulation
- Consistent error handling

### 3. Middleware Pattern

**Purpose**: Cross-cutting concerns without cluttering business logic.

```go
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := time.Since(start).Seconds()
        RecordHTTPRequest(method, path, statusCode, duration)
    }
}
```

**Benefits**:
- Separation of concerns
- Reusable across routes
- Clean request/response handling

### 4. Factory Pattern

**Purpose**: Create instances with proper configuration.

```go
func NewGrpcUserService() *GrpcUserService {
    return &GrpcUserService{
        userService: &service.UserService{},
    }
}
```

**Benefits**:
- Encapsulation of creation logic
- Consistent initialization
- Dependency injection

## 🚀 Go-Specific Learnings

### 1. Package Management

```go
// go.mod - Module definition
module github.com/114windd/restapi

// Import paths
import "github.com/114windd/restapi/internal/service"
import "github.com/114windd/restapi/pkg/models"
```

**Key Points**:
- Module path should match repository URL
- `internal/` packages can't be imported by external projects
- `pkg/` packages are public and can be imported

### 2. Interface Design

```go
type UserService interface {
    CreateUser(name, email, password string) (*models.User, error)
    GetUser(id uint) (*models.User, error)
    // ...
}
```

**Key Points**:
- Small, focused interfaces
- Interface segregation principle
- Easy to mock for testing

### 3. Error Handling

```go
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return nil, status.Error(codes.AlreadyExists, "email already exists")
    }
    return nil, status.Error(codes.Internal, "failed to create user")
}
```

**Key Points**:
- Explicit error handling
- Error wrapping for context
- Different error types for different responses

### 4. Context Usage

```go
func (s *GrpcUserService) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.UserResponse, error) {
    // Use context for cancellation, timeouts, etc.
}
```

**Key Points**:
- Context for request-scoped values
- Cancellation and timeout handling
- Request tracing

## 🔍 Testing Strategies

### 1. Unit Testing
- Test individual functions in isolation
- Mock dependencies
- Focus on business logic

### 2. Integration Testing
- Test API endpoints
- Test database interactions
- Test service integration

### 3. End-to-End Testing
- Test complete user workflows
- Test both REST and gRPC interfaces
- Test error scenarios

## 📊 Monitoring & Observability

### 1. Metrics
- **Counters**: Request counts, error counts
- **Histograms**: Request duration, response sizes
- **Gauges**: Current state (health status)

### 2. Logging
- Structured logging with context
- Different log levels
- Correlation IDs for tracing

### 3. Health Checks
- Database connectivity
- Service dependencies
- Readiness and liveness probes

## 🐳 Containerization

### 1. Multi-stage Dockerfile
```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
# ... build steps

# Final stage
FROM alpine:latest
# ... runtime steps
```

### 2. Docker Compose
- Service orchestration
- Environment configuration
- Volume management

## 🎯 Key Takeaways

### 1. Architecture
- **Separation of Concerns**: Each layer has a single responsibility
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Interface Segregation**: Small, focused interfaces

### 2. Go Best Practices
- **Package Organization**: Clear structure with `cmd/`, `internal/`, `pkg/`
- **Error Handling**: Explicit and contextual
- **Concurrency**: Use goroutines for parallel processing

### 3. Production Readiness
- **Monitoring**: Comprehensive metrics and logging
- **Resilience**: Retry patterns and circuit breakers
- **Security**: Authentication, input validation, secure defaults

### 4. API Design
- **Consistency**: Similar patterns across REST and gRPC
- **Documentation**: Clear API contracts
- **Versioning**: Plan for API evolution

### 5. Development Workflow
- **Makefile**: Automate common tasks
- **Scripts**: Test and deployment automation
- **Documentation**: Keep it up to date

## 🔗 Further Learning

### 1. Advanced Topics
- **Circuit Breakers**: Prevent cascading failures
- **Distributed Tracing**: Request flow across services
- **Service Mesh**: Istio, Linkerd for microservices
- **Event Sourcing**: Event-driven architecture

### 2. Go-Specific
- **Go Modules**: Advanced dependency management
- **Go Generics**: Type-safe generic programming
- **Go Profiling**: Performance optimization
- **Go Testing**: Advanced testing techniques

### 3. Microservices
- **API Gateway**: Kong, Ambassador
- **Service Discovery**: Consul, etcd
- **Message Queues**: RabbitMQ, Apache Kafka
- **CQRS**: Command Query Responsibility Segregation

This project demonstrates many fundamental concepts in modern software development, from clean architecture to production monitoring. The patterns and techniques used here are applicable to many other projects and technologies.

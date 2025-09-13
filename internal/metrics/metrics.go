package metrics

import (
	"context"
	"time"

	"github.com/114windd/restapi/internal/database"
	"github.com/114windd/restapi/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// Prometheus metrics
var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// gRPC metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status_code"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Database metrics
	dbOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)

	dbOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Health check metrics
	healthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "health_check_status",
			Help: "Health check status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service"},
	)
)

// PrometheusMiddleware creates a Gin middleware for Prometheus metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()

		RecordHTTPRequest(method, path, statusCode, duration)
	}
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path string, statusCode int, duration float64) {
	httpRequestsTotal.WithLabelValues(method, path, string(rune(statusCode))).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// GrpcPrometheusInterceptor creates a gRPC interceptor for Prometheus metrics
func GrpcPrometheusInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		method := info.FullMethod

		// Process request
		resp, err := handler(ctx, req)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := grpc.Code(err).String()

		grpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
		grpcRequestDuration.WithLabelValues(method).Observe(duration)

		return resp, err
	}
}

// RecordDatabaseOperation records metrics for database operations
func RecordDatabaseOperation(operation, table, status string, duration time.Duration) {
	dbOperationsTotal.WithLabelValues(operation, table, status).Inc()
	dbOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateHealthStatus updates the health check status metric
func UpdateHealthStatus(service string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	healthCheckStatus.WithLabelValues(service).Set(status)
}

// SetupMetricsRoutes sets up the /metrics endpoint
func SetupMetricsRoutes(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// HealthCheckHandler handles the /healthz endpoint
func HealthCheckHandler(c *gin.Context) {
	// Check database connectivity
	start := time.Now()
	err := database.GetDB().Exec("SELECT 1").Error
	duration := time.Since(start)

	healthy := err == nil
	UpdateHealthStatus("database", healthy)

	if healthy {
		RecordDatabaseOperation("health_check", "users", "success", duration)
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  "connected",
		})
	} else {
		RecordDatabaseOperation("health_check", "users", "error", duration)
		logger.Log.Error("Health check failed - database unreachable", "error", err)
		c.JSON(500, gin.H{
			"status":    "unhealthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  "disconnected",
			"error":     err.Error(),
		})
	}
}

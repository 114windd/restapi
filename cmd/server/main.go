package main

import (
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/114windd/restapi/internal/api"
	"github.com/114windd/restapi/internal/database"
	grpcserver "github.com/114windd/restapi/internal/grpc"
	"github.com/114windd/restapi/internal/logger"
	"github.com/114windd/restapi/internal/metrics"
	"github.com/114windd/restapi/pkg/proto"
)

func main() {
	// Initialize logger first
	logger.Init()
	logger.Log.Info("Starting hybrid REST + gRPC API server")

	// Initialize database
	database.InitDB()

	// Start gRPC server in a goroutine
	go startGrpcServer()

	// Setup Gin router with logging and metrics middleware
	r := gin.New()
	r.Use(api.LoggingMiddleware())
	r.Use(metrics.PrometheusMiddleware())
	r.Use(gin.Recovery())

	// Health check and metrics routes
	r.GET("/healthz", metrics.HealthCheckHandler)
	metrics.SetupMetricsRoutes(r)

	// Public routes
	r.POST("/signup", api.Signup)
	r.POST("/login", api.Login)

	// Protected routes
	protected := r.Group("/")
	protected.Use(api.AuthMiddleware())
	{
		protected.GET("/users", api.GetUsers)
		protected.GET("/users/:id", api.GetUser)
		protected.PUT("/users/:id", api.UpdateUser)
		protected.DELETE("/users/:id", api.DeleteUser)
	}

	logger.Log.Info("REST server starting on :8080")
	logger.Log.Info("gRPC server starting on :50051")
	logger.Log.Info("Metrics available at :8080/metrics")
	logger.Log.Info("Health check available at :8080/healthz")

	if err := r.Run(":8080"); err != nil {
		logger.Log.WithError(err).Fatal("Failed to start REST server")
	}
}

// startGrpcServer starts the gRPC server
func startGrpcServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to listen on :50051")
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.GrpcPrometheusInterceptor()),
	)

	// Register the user service
	userService := grpcserver.NewGrpcUserService()
	proto.RegisterUserServiceServer(grpcServer, userService)

	logger.Log.Info("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.WithError(err).Fatal("Failed to serve gRPC")
	}
}

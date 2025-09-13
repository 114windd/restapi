package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User model
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // "-" excludes from JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Request structs
type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	db        *gorm.DB
	jwtSecret = []byte("mock-secret-key")
)

func main() {
	// Initialize logger first
	Init()
	Log.Info("Starting REST API server")

	// Initialize database
	initDB()

	// Setup Gin router with logging middleware
	r := gin.New()
	r.Use(loggingMiddleware())
	r.Use(gin.Recovery())

	// Public routes
	r.POST("/signup", signup)
	r.POST("/login", login)

	// Protected routes
	protected := r.Group("/")
	protected.Use(authMiddleware())
	{
		protected.GET("/users", getUsers)
		protected.GET("/users/:id", getUser)
		protected.PUT("/users/:id", updateUser)
		protected.DELETE("/users/:id", deleteUser)
	}

	Log.Info("Server starting on :8080")
	r.Run(":8080")
}

func initDB() {
	// Database connection string
	dsn := "host=localhost user=postgres password=postgres dbname=restapi port=5432 sslmode=disable"

	// You can override with environment variable
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		Log.WithError(err).Fatal("Failed to connect to database")
	}

	// Auto-migrate the schema
	LogDatabase("migrate", "users").Info("Running database migration")
	err = db.AutoMigrate(&User{})
	if err != nil {
		Log.WithError(err).Fatal("Failed to migrate database")
	}

	Log.Info("Database connected and migrated successfully")
}

// Logging middleware
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log after processing
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		entry := LogRequest(method, path, getUserIDFromContext(c))
		entry = entry.WithFields(map[string]interface{}{
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"client_ip":   c.ClientIP(),
		})

		if statusCode >= 400 {
			entry.Warn("Request completed with error")
		} else {
			entry.Info("Request completed successfully")
		}
	}
}

// Helper to get user ID from context for logging
func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return strconv.Itoa(int(userID.(uint)))
	}
	return "anonymous"
}

// JWT helper functions
func generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Log.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			Log.WithError(err).Warn("Invalid JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["user_id"].(float64))
		c.Set("user_id", userID)
		c.Next()
	}
}

// Auth handlers
func signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Log.WithError(err).Warn("Invalid signup request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	LogAuth("signup_attempt", req.Email).Info("User signup attempt")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Log.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := CreateUserWithRetry(&user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT
	token, err := generateJWT(user.ID)
	if err != nil {
		Log.WithError(err).Error("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	LogAuth("signup_success", req.Email).WithField("user_id", user.ID).Info("User created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
		"token":   token,
	})
}

func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Log.WithError(err).Warn("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	LogAuth("login_attempt", req.Email).Info("User login attempt")

	// Find user by email with retry
	user, err := FindUserByEmailWithRetry(req.Email)
	if err != nil {
		LogAuth("login_failed", req.Email).Warn("User not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		LogAuth("login_failed", req.Email).Warn("Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	token, err := generateJWT(user.ID)
	if err != nil {
		Log.WithError(err).Error("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	LogAuth("login_success", req.Email).WithField("user_id", user.ID).Info("User logged in successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// CRUD handlers
func getUsers(c *gin.Context) {
	users, err := GetAllUsersWithRetry()
	if err != nil {
		LogDatabase("select", "users").WithError(err).Error("Failed to fetch users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	LogDatabase("select", "users").WithField("count", len(users)).Info("Users fetched successfully")
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func getUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithError(err).Warn("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := FindUserByIDWithRetry(uint(id))
	if err != nil {
		LogDatabase("select", "users").WithField("user_id", id).Warn("User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	LogDatabase("select", "users").WithField("user_id", id).Info("User fetched successfully")
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithError(err).Warn("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Log.WithError(err).Warn("Invalid update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := FindUserByIDWithRetry(uint(id))
	if err != nil {
		LogDatabase("select", "users").WithField("user_id", id).Warn("User not found for update")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := UpdateUserWithRetry(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	LogDatabase("update", "users").WithField("user_id", id).Info("User updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithError(err).Warn("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := DeleteUserWithRetry(uint(id)); err != nil {
		LogDatabase("delete", "users").WithError(err).WithField("user_id", id).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	LogDatabase("delete", "users").WithField("user_id", id).Info("User deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

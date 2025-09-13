package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/114windd/restapi/internal/logger"
	"github.com/114windd/restapi/internal/service"
	"github.com/114windd/restapi/pkg/models"
)

var (
	jwtSecret = []byte("mock-secret-key")
)

// Auth handlers
func Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.WithError(err).Warn("Invalid signup request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.LogAuth("signup_attempt", req.Email).Info("User signup attempt")

	// Use the service layer
	user, err := service.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
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
		logger.Log.WithError(err).Error("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	logger.LogAuth("signup_success", req.Email).WithField("user_id", user.ID).Info("User created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
		"token":   token,
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.WithError(err).Warn("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.LogAuth("login_attempt", req.Email).Info("User login attempt")

	// Use the service layer
	user, err := service.GetUserByEmail(req.Email)
	if err != nil {
		logger.LogAuth("login_failed", req.Email).Warn("User not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := service.ValidatePassword(user, req.Password); err != nil {
		logger.LogAuth("login_failed", req.Email).Warn("Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	token, err := generateJWT(user.ID)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	logger.LogAuth("login_success", req.Email).WithField("user_id", user.ID).Info("User logged in successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// CRUD handlers
func GetUsers(c *gin.Context) {
	users, err := service.ListUsers()
	if err != nil {
		logger.LogDatabase("select", "users").WithError(err).Error("Failed to fetch users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	logger.LogDatabase("select", "users").WithField("count", len(users)).Info("Users fetched successfully")
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.WithError(err).Warn("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := service.GetUser(uint(id))
	if err != nil {
		logger.LogDatabase("select", "users").WithField("user_id", id).Warn("User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	logger.LogDatabase("select", "users").WithField("user_id", id).Info("User fetched successfully")
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.WithError(err).Warn("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.RestUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.WithError(err).Warn("Invalid update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := service.UpdateUser(uint(id), req.Name, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	logger.LogDatabase("update", "users").WithField("user_id", id).Info("User updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.WithError(err).Warn("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := service.DeleteUser(uint(id)); err != nil {
		logger.LogDatabase("delete", "users").WithError(err).WithField("user_id", id).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	logger.LogDatabase("delete", "users").WithField("user_id", id).Info("User deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Log.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			logger.Log.WithError(err).Warn("Invalid JWT token")
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

// Helper to get user ID from context for logging
func GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return strconv.Itoa(int(userID.(uint)))
	}
	return "anonymous"
}

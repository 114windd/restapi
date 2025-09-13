package database

import (
	"errors"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/114windd/restapi/internal/logger"
	"github.com/114windd/restapi/internal/retry"
	"github.com/114windd/restapi/pkg/models"
)

var db *gorm.DB

// InitDB initializes the database connection
func InitDB() {
	// Database connection string
	dsn := "host=localhost user=postgres password=postgres dbname=restapi port=5432 sslmode=disable"

	// You can override with environment variable
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to connect to database")
	}

	// Auto-migrate the schema
	logger.LogDatabase("migrate", "users").Info("Running database migration")
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to migrate database")
	}

	logger.Log.Info("Database connected and migrated successfully")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// Database operations with retry logic

// CreateUserWithRetry creates a user with retry logic
func CreateUserWithRetry(user *models.User) error {
	config := retry.DefaultRetryConfig()

	err := retry.ExecuteWithRetry("create_user", func() error {
		logger.LogDatabase("create", "users").WithField("email", user.Email).Debug("Attempting to create user")

		err := db.Create(user).Error
		if err != nil {
			// Don't retry on unique constraint violations (business logic errors)
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
				logger.LogDatabase("create", "users").WithError(err).Warn("Unique constraint violation - not retrying")
				return err // Return immediately, don't retry
			}
		}
		return err
	}, config)

	// Metrics recording moved to service layer

	return err
}

// FindUserByEmailWithRetry finds a user by email with retry logic
func FindUserByEmailWithRetry(email string) (*models.User, error) {
	var user models.User
	config := retry.DefaultRetryConfig()

	err := retry.ExecuteWithRetry("find_user_by_email", func() error {
		logger.LogDatabase("select", "users").WithField("email", email).Debug("Attempting to find user by email")

		err := db.Where("email = ?", email).First(&user).Error
		if err != nil {
			// Don't retry on "not found" errors (business logic errors)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.LogDatabase("select", "users").WithField("email", email).Debug("User not found - not retrying")
				return err // Return immediately, don't retry
			}
		}
		return err
	}, config)

	// Metrics recording moved to service layer

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByIDWithRetry finds a user by ID with retry logic
func FindUserByIDWithRetry(id uint) (*models.User, error) {
	var user models.User
	config := retry.DefaultRetryConfig()

	err := retry.ExecuteWithRetry("find_user_by_id", func() error {
		logger.LogDatabase("select", "users").WithField("user_id", id).Debug("Attempting to find user by ID")

		err := db.First(&user, id).Error
		if err != nil {
			// Don't retry on "not found" errors
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.LogDatabase("select", "users").WithField("user_id", id).Debug("User not found - not retrying")
				return err
			}
		}
		return err
	}, config)

	// Metrics recording moved to service layer

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserWithRetry updates a user with retry logic
func UpdateUserWithRetry(user *models.User) error {
	config := retry.DefaultRetryConfig()

	err := retry.ExecuteWithRetry("update_user", func() error {
		logger.LogDatabase("update", "users").WithField("user_id", user.ID).Debug("Attempting to update user")

		err := db.Save(user).Error
		if err != nil {
			// Don't retry on unique constraint violations
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
				logger.LogDatabase("update", "users").WithError(err).Warn("Unique constraint violation - not retrying")
				return err
			}
		}
		return err
	}, config)

	// Metrics recording moved to service layer

	return err
}

// DeleteUserWithRetry deletes a user with retry logic
func DeleteUserWithRetry(id uint) error {
	config := retry.DefaultRetryConfig()

	err := retry.ExecuteWithRetry("delete_user", func() error {
		logger.LogDatabase("delete", "users").WithField("user_id", id).Debug("Attempting to delete user")

		return db.Delete(&models.User{}, id).Error
	}, config)

	// Metrics recording moved to service layer

	return err
}

// GetAllUsersWithRetry gets all users with retry logic
func GetAllUsersWithRetry() ([]models.User, error) {
	var users []models.User
	config := retry.DefaultRetryConfig()

	err := retry.ExecuteWithRetry("get_all_users", func() error {
		logger.LogDatabase("select", "users").Debug("Attempting to fetch all users")

		return db.Find(&users).Error
	}, config)

	// Metrics recording moved to service layer

	if err != nil {
		return nil, err
	}
	return users, nil
}

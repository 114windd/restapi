package main

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Database operations with retry logic

// CreateUserWithRetry creates a user with retry logic
func CreateUserWithRetry(user *User) error {
	config := DefaultRetryConfig()

	return ExecuteWithRetry("create_user", func() error {
		LogDatabase("create", "users").WithField("email", user.Email).Debug("Attempting to create user")

		err := db.Create(user).Error
		if err != nil {
			// Don't retry on unique constraint violations (business logic errors)
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
				LogDatabase("create", "users").WithError(err).Warn("Unique constraint violation - not retrying")
				return err // Return immediately, don't retry
			}
		}
		return err
	}, config)
}

// FindUserByEmailWithRetry finds a user by email with retry logic
func FindUserByEmailWithRetry(email string) (*User, error) {
	var user User
	config := DefaultRetryConfig()

	err := ExecuteWithRetry("find_user_by_email", func() error {
		LogDatabase("select", "users").WithField("email", email).Debug("Attempting to find user by email")

		err := db.Where("email = ?", email).First(&user).Error
		if err != nil {
			// Don't retry on "not found" errors (business logic errors)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				LogDatabase("select", "users").WithField("email", email).Debug("User not found - not retrying")
				return err // Return immediately, don't retry
			}
		}
		return err
	}, config)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByIDWithRetry finds a user by ID with retry logic
func FindUserByIDWithRetry(id uint) (*User, error) {
	var user User
	config := DefaultRetryConfig()

	err := ExecuteWithRetry("find_user_by_id", func() error {
		LogDatabase("select", "users").WithField("user_id", id).Debug("Attempting to find user by ID")

		err := db.First(&user, id).Error
		if err != nil {
			// Don't retry on "not found" errors
			if errors.Is(err, gorm.ErrRecordNotFound) {
				LogDatabase("select", "users").WithField("user_id", id).Debug("User not found - not retrying")
				return err
			}
		}
		return err
	}, config)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserWithRetry updates a user with retry logic
func UpdateUserWithRetry(user *User) error {
	config := DefaultRetryConfig()

	return ExecuteWithRetry("update_user", func() error {
		LogDatabase("update", "users").WithField("user_id", user.ID).Debug("Attempting to update user")

		err := db.Save(user).Error
		if err != nil {
			// Don't retry on unique constraint violations
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
				LogDatabase("update", "users").WithError(err).Warn("Unique constraint violation - not retrying")
				return err
			}
		}
		return err
	}, config)
}

// DeleteUserWithRetry deletes a user with retry logic
func DeleteUserWithRetry(id uint) error {
	config := DefaultRetryConfig()

	return ExecuteWithRetry("delete_user", func() error {
		LogDatabase("delete", "users").WithField("user_id", id).Debug("Attempting to delete user")

		return db.Delete(&User{}, id).Error
	}, config)
}

// GetAllUsersWithRetry gets all users with retry logic
func GetAllUsersWithRetry() ([]User, error) {
	var users []User
	config := DefaultRetryConfig()

	err := ExecuteWithRetry("get_all_users", func() error {
		LogDatabase("select", "users").Debug("Attempting to fetch all users")

		return db.Find(&users).Error
	}, config)

	if err != nil {
		return nil, err
	}
	return users, nil
}

package service

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/114windd/restapi/internal/database"
	"github.com/114windd/restapi/pkg/models"
)

// UserService contains shared business logic
type UserService struct{}

// CreateUser creates a new user
func (s *UserService) CreateUser(name, email, password string) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := database.CreateUserWithRetry(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id uint) (*models.User, error) {
	return database.FindUserByIDWithRetry(id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return database.FindUserByEmailWithRetry(email)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id uint, name, email string) (*models.User, error) {
	user, err := database.FindUserByIDWithRetry(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}

	if err := database.UpdateUserWithRetry(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id uint) error {
	return database.DeleteUserWithRetry(id)
}

// ListUsers returns all users
func (s *UserService) ListUsers() ([]models.User, error) {
	return database.GetAllUsersWithRetry()
}

// ValidatePassword checks if password is correct
func (s *UserService) ValidatePassword(user *models.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

// Global service instance
var userService = &UserService{}

// Package-level functions for easy access
func CreateUser(name, email, password string) (*models.User, error) {
	return userService.CreateUser(name, email, password)
}

func GetUser(id uint) (*models.User, error) {
	return userService.GetUser(id)
}

func GetUserByEmail(email string) (*models.User, error) {
	return userService.GetUserByEmail(email)
}

func UpdateUser(id uint, name, email string) (*models.User, error) {
	return userService.UpdateUser(id, name, email)
}

func DeleteUser(id uint) error {
	return userService.DeleteUser(id)
}

func ListUsers() ([]models.User, error) {
	return userService.ListUsers()
}

func ValidatePassword(user *models.User, password string) error {
	return userService.ValidatePassword(user, password)
}

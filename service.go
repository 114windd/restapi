package main

import (
	"golang.org/x/crypto/bcrypt"
)

// UserService contains shared business logic
type UserService struct{}

// CreateUser creates a new user
func (s *UserService) CreateUser(name, email, password string) (*User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := CreateUserWithRetry(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id uint) (*User, error) {
	return FindUserByIDWithRetry(id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*User, error) {
	return FindUserByEmailWithRetry(email)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id uint, name, email string) (*User, error) {
	user, err := FindUserByIDWithRetry(id)
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

	if err := UpdateUserWithRetry(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id uint) error {
	return DeleteUserWithRetry(id)
}

// ListUsers returns all users
func (s *UserService) ListUsers() ([]User, error) {
	return GetAllUsersWithRetry()
}

// ValidatePassword checks if password is correct
func (s *UserService) ValidatePassword(user *User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

// Global service instance
var userService = &UserService{}

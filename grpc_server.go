package main

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcUserService implements the gRPC UserService
type GrpcUserService struct {
	UnimplementedUserServiceServer
	userService *UserService
}

// NewGrpcUserService creates a new gRPC user service
func NewGrpcUserService() *GrpcUserService {
	return &GrpcUserService{
		userService: userService,
	}
}

// CreateUser implements the CreateUser gRPC method
func (s *GrpcUserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	Log.Info("gRPC CreateUser request", "email", req.Email, "name", req.Name)

	// Validate request
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email, and password are required")
	}

	// Use the existing UserService
	user, err := s.userService.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			Log.Warn("gRPC CreateUser failed - email already exists", "email", req.Email)
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		Log.Error("gRPC CreateUser failed", "error", err, "email", req.Email)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Convert to ProtoUser
	protoUser := &ProtoUser{
		Id:        uint32(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	Log.Info("gRPC CreateUser success", "user_id", user.ID, "email", req.Email)
	return &UserResponse{
		User:    protoUser,
		Message: "User created successfully",
	}, nil
}

// GetUser implements the GetUser gRPC method
func (s *GrpcUserService) GetUser(ctx context.Context, req *GetUserRequest) (*UserResponse, error) {
	Log.Info("gRPC GetUser request", "user_id", req.Id)

	// Use the existing UserService
	user, err := s.userService.GetUser(uint(req.Id))
	if err != nil {
		Log.Warn("gRPC GetUser failed - user not found", "user_id", req.Id)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Convert to ProtoUser
	protoUser := &ProtoUser{
		Id:        uint32(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	Log.Info("gRPC GetUser success", "user_id", req.Id)
	return &UserResponse{
		User:    protoUser,
		Message: "User retrieved successfully",
	}, nil
}

// UpdateUser implements the UpdateUser gRPC method
func (s *GrpcUserService) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UserResponse, error) {
	Log.Info("gRPC UpdateUser request", "user_id", req.Id, "name", req.Name, "email", req.Email)

	// Use the existing UserService
	user, err := s.userService.UpdateUser(uint(req.Id), req.Name, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			Log.Warn("gRPC UpdateUser failed - email already exists", "user_id", req.Id, "email", req.Email)
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		Log.Error("gRPC UpdateUser failed", "error", err, "user_id", req.Id)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	// Convert to ProtoUser
	protoUser := &ProtoUser{
		Id:        uint32(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	Log.Info("gRPC UpdateUser success", "user_id", req.Id)
	return &UserResponse{
		User:    protoUser,
		Message: "User updated successfully",
	}, nil
}

// DeleteUser implements the DeleteUser gRPC method
func (s *GrpcUserService) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	Log.Info("gRPC DeleteUser request", "user_id", req.Id)

	// Use the existing UserService
	err := s.userService.DeleteUser(uint(req.Id))
	if err != nil {
		Log.Error("gRPC DeleteUser failed", "error", err, "user_id", req.Id)
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	Log.Info("gRPC DeleteUser success", "user_id", req.Id)
	return &DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

// ListUsers implements the ListUsers gRPC method
func (s *GrpcUserService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	Log.Info("gRPC ListUsers request")

	// Use the existing UserService
	users, err := s.userService.ListUsers()
	if err != nil {
		Log.Error("gRPC ListUsers failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	// Convert to ProtoUser slice
	protoUsers := make([]*ProtoUser, len(users))
	for i, user := range users {
		protoUsers[i] = &ProtoUser{
			Id:        uint32(user.ID),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		}
	}

	Log.Info("gRPC ListUsers success", "count", len(users))
	return &ListUsersResponse{
		Users: protoUsers,
	}, nil
}

// Helper function to convert User to ProtoUser
func userToProtoUser(user *User) *ProtoUser {
	return &ProtoUser{
		Id:        uint32(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

// Helper function to convert ProtoUser to User (if needed)
func protoUserToUser(protoUser *ProtoUser) *User {
	createdAt, _ := time.Parse(time.RFC3339, protoUser.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, protoUser.UpdatedAt)

	return &User{
		ID:        uint(protoUser.Id),
		Name:      protoUser.Name,
		Email:     protoUser.Email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

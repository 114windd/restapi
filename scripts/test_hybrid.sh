#!/bin/bash

# Test script for hybrid REST + gRPC service

echo "Testing Hybrid REST + gRPC Service"
echo "=================================="

# Test health check
echo "1. Testing health check endpoint..."
curl -s http://localhost:8080/healthz | jq '.' || echo "Health check failed"

echo -e "\n2. Testing metrics endpoint..."
curl -s http://localhost:8080/metrics | head -20

echo -e "\n3. Testing REST API - Signup..."
curl -s -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}' | jq '.'

echo -e "\n4. Testing REST API - Login..."
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq '.'

echo -e "\n5. Testing REST API - Get Users (requires auth)..."
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

if [ "$TOKEN" != "null" ]; then
  curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/users | jq '.'
else
  echo "Failed to get token for authenticated request"
fi

echo -e "\n6. Testing gRPC service (requires grpcurl)..."
if command -v grpcurl &> /dev/null; then
  echo "Testing gRPC CreateUser..."
  grpcurl -plaintext -d '{"name":"gRPC User","email":"grpc@example.com","password":"password123"}' \
    localhost:50051 user.UserService/CreateUser
  
  echo -e "\nTesting gRPC ListUsers..."
  grpcurl -plaintext localhost:50051 user.UserService/ListUsers
else
  echo "grpcurl not installed. Install with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

echo -e "\nTest completed!"

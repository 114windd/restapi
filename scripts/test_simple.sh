#!/bin/bash

# Simple test script for hybrid REST + gRPC service (no external dependencies)

echo "Testing Hybrid REST + gRPC Service"
echo "=================================="

# Check if service is running
echo "1. Checking if service is running..."
if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
    echo "✅ Service is running"
else
    echo "❌ Service is not running. Please start it with: ./hybrid-api"
    exit 1
fi

# Test health check
echo -e "\n2. Testing health check endpoint..."
echo "Response:"
curl -s http://localhost:8080/healthz
echo -e "\n"

# Test metrics endpoint
echo -e "\n3. Testing metrics endpoint..."
echo "First 10 lines of metrics:"
curl -s http://localhost:8080/metrics | head -10
echo -e "\n"

# Test REST API - Signup
echo -e "\n4. Testing REST API - Signup..."
echo "Request: POST /signup"
echo "Response:"
curl -s -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'
echo -e "\n"

# Test REST API - Login
echo -e "\n5. Testing REST API - Login..."
echo "Request: POST /login"
echo "Response:"
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
echo -e "\n"

# Test REST API - Get Users (with auth)
echo -e "\n6. Testing REST API - Get Users (requires auth)..."
echo "Getting token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}')

echo "Login response: $TOKEN_RESPONSE"

# Extract token using basic text processing (no jq needed)
TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo "Token extracted: ${TOKEN:0:20}..."
    echo "Request: GET /users"
    echo "Response:"
    curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/users
    echo -e "\n"
else
    echo "❌ Failed to get token for authenticated request"
fi

# Test gRPC service
echo -e "\n7. Testing gRPC service..."
if command -v grpcurl &> /dev/null; then
    echo "✅ grpcurl found, testing gRPC..."
    echo "Testing gRPC CreateUser..."
    grpcurl -plaintext -d '{"name":"gRPC User","email":"grpc@example.com","password":"password123"}' \
      localhost:50051 user.UserService/CreateUser
    
    echo -e "\nTesting gRPC ListUsers..."
    grpcurl -plaintext localhost:50051 user.UserService/ListUsers
else
    echo "❌ grpcurl not installed"
    echo "To install grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    echo "Or install jq for better JSON parsing: sudo apt install jq"
fi

echo -e "\n✅ Test completed!"
echo -e "\nTo install missing tools:"
echo "  jq: sudo apt install jq"
echo "  grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"

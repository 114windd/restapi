# Go REST API

A simple REST API built with Go, featuring user authentication and CRUD operations.

## Features

- User registration and login
- JWT-based authentication
- Password hashing with bcrypt
- PostgreSQL database integration
- CRUD operations for users
- Input validation
- Protected routes

## Tech Stack

- **Go** - Programming language
- **Gin** - HTTP web framework
- **GORM** - ORM for database operations
- **PostgreSQL** - Database
- **JWT** - Authentication tokens
- **Bcrypt** - Password hashing

## Prerequisites

- Go 1.19 or higher
- PostgreSQL database
- Git

## Installation

1. Clone the repository:
```bash
git clone <your-repo-url>
cd go-rest-api
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up PostgreSQL database:
   - Create a database named `restapi`
   - Update connection string in `main.go` if needed

4. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Public Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/signup` | Register a new user |
| POST | `/login` | Login user |

### Protected Routes (Requires JWT token)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | Get all users |
| GET | `/users/:id` | Get user by ID |
| PUT | `/users/:id` | Update user |
| DELETE | `/users/:id` | Delete user |

## Usage Examples

### Register a new user
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Get all users (with token)
```bash
curl -X GET http://localhost:8080/users \
  -H "Authorization: Bearer <your-jwt-token>"
```

### Update user
```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com"
  }'
```

## Environment Variables

You can override the default database connection using:

```bash
export DATABASE_URL="host=localhost user=your_user password=your_password dbname=your_db port=5432 sslmode=disable"
```

## Database Schema

The application automatically creates the following table:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Security Features

- Passwords are hashed using bcrypt
- JWT tokens expire after 24 hours
- Password fields are excluded from JSON responses
- Email uniqueness is enforced
- Input validation on all endpoints

## Project Structure

```
.
├── main.go          # Main application file
├── go.mod           # Go module file
├── go.sum           # Go dependencies
└── README.md        # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is open source and available under the [MIT License](LICENSE).
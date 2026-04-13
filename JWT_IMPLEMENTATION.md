# JWT Token Implementation & Authorization

## Overview
This document explains the JWT (JSON Web Token) implementation for authentication in the UserService backend.

## What Was Changed

### 1. JWT Token Utility (`handler/jwt.go`) - NEW FILE
- **GenerateToken()**: Creates a signed JWT token with 24-hour expiration
- **ValidateToken()**: Validates and extracts user ID from JWT token
- Uses HMAC-SHA256 signing method for token security

**Key Features:**
- Tokens include user ID and standard JWT claims (IssuedAt, ExpiresAt, Subject)
- 24-hour token validity period
- Proper error handling for expired/invalid tokens

### 2. Server Configuration (`handler/server.go`)
- Added `JWTSecret` field to Server struct
- Updated `NewServerOptions` to accept JWT secret during initialization

### 3. Authentication Middleware (`handler/middleware.go`)
- Updated `BearerTokenMiddleware()` to validate JWT tokens instead of placeholder format
- Removed legacy placeholder token extraction function
- Validates JWT signature and expiration
- Extracts user ID from valid JWT and stores in Echo context

**Middleware Behavior:**
- Validates `Authorization: Bearer <token>` header format
- Checks JWT signature and expiration
- Returns 401 Unauthorized with descriptive error message on invalid tokens

### 4. Login Endpoint (`handler/endpoints.go`)
- Updated `PostLogin()` to generate JWT tokens via `GenerateToken()`
- Replaces placeholder token generation with real JWT tokens
- Returns signed JWT token to client on successful login

### 5. Main Application (`cmd/main.go`)
- Reads `JWT_SECRET` from environment variables
- Falls back to default secret if not provided (for development)
- Passes JWT secret to server initialization

**⚠️ IMPORTANT**: In production, set the `JWT_SECRET` environment variable to a strong, random secret!

## Public vs Protected Routes

### Public Routes (No Authorization Required)
- `GET /hello` - Test endpoint
- `POST /login` - User login
- `POST /users` - User registration

### Protected Routes (JWT Token Required)
- `POST /estate` - Create estate
- `POST /estate/{id}/tree` - Add tree to estate
- `GET /estate/{id}/stats` - Get estate statistics
- `GET /estate/{id}/drone-plan` - Get drone survey plan
- `GET /users/{id}` - Get user profile
- `PUT /users/{id}` - Update user profile
- `DELETE /users/{id}` - Delete user account

## How to Use

### 1. Obtain a Token
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

Response:
```json
{
  "Token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzNDU2Nzg5MCIsImV4cCI6MTcxMjAxMzI0OX0..."
}
```

### 2. Use Token for Protected Routes
```bash
curl -X POST http://localhost:8080/estate \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "length": 100,
    "width": 100
  }'
```

## Environment Configuration

### Required Environment Variables
```bash
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your-super-secret-key-change-in-production
```

### Development Setup
```bash
# Using Docker Compose
export JWT_SECRET="dev-secret-key"
docker compose up --build

# Or running locally
export JWT_SECRET="dev-secret-key"
go run ./cmd
```

## Token Structure

JWT tokens contain:
- **Header**: Algorithm (HS256) and token type (JWT)
- **Payload (Claims)**:
  - `user_id`: The UUID of the authenticated user
  - `sub`: JWT standard claim (same as user_id)
  - `iat`: Issued at timestamp
  - `exp`: Expiration timestamp (24 hours from issuance)
- **Signature**: HMAC-SHA256 signed with JWT_SECRET

## Error Responses

### Missing Authorization Header
```json
{
  "Message": "Missing authorization header"
}
```

### Invalid Header Format
```json
{
  "Message": "Invalid authorization header format"
}
```

### Invalid or Expired Token
```json
{
  "Message": "Invalid or expired token: token has expired"
}
```

## Security Notes

1. **Secret Management**: Store JWT_SECRET securely (environment variables, secret manager, etc.)
2. **HTTPS Only**: Always use HTTPS in production to prevent token interception
3. **Token Expiration**: Tokens expire after 24 hours - users must re-login
4. **Signature Verification**: All tokens are validated on each protected request
5. **No Token Refresh**: Current implementation doesn't support refresh tokens

## Testing authenticated Endpoints

Use the Postman collection with the following pattern:
1. Call `/login` endpoint and copy the token from response
2. Add `Authorization: Bearer <token>` header to subsequent requests
3. Protected endpoints will extract user ID from the JWT

---

**Implementation Date**: April 13, 2026
**JWT Library**: github.com/golang-jwt/jwt/v4

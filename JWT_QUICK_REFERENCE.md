# JWT Implementation - Quick Reference Guide

## Summary of Changes

✅ **JWT Token Generation** - Replaced placeholder tokens with signed JWT tokens
✅ **JWT Validation** - Middleware validates JWT signature and expiration
✅ **24-hour Token Expiry** - Tokens expire after 24 hours, requiring re-login
✅ **HMAC-SHA256** - Using secure signing method
✅ **All Tests Passing** - 100% test coverage for JWT authentication

---

## Files Modified

| File | Change |
|------|--------|
| `handler/jwt.go` | **NEW** - JWT generation and validation |
| `handler/server.go` | Added `JWTSecret` field to Server struct |
| `handler/middleware.go` | Updated to validate JWT tokens |
| `handler/endpoints.go` | Login endpoint returns JWT tokens |
| `cmd/main.go` | Reads JWT_SECRET from environment |
| `handler/endpoints_test.go` | Updated tests to use JWT tokens |
| `go.mod` | Added `github.com/golang-jwt/jwt/v4` dependency |

---

## Usage Examples

### 1. Login and Get Token
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myuser",
    "password": "mypassword"
  }'

# Response:
# {
#   "Token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzLWQ0NWQtNDVkMS1iNGM1LWU5YmQ3ZTg2YmRkMSIsImV4cCI6MTcxMjA1ODMyM30..."
# }
```

### 2. Use Token for Protected Endpoints
```bash
# Creating an estate (protected)
curl -X POST http://localhost:8080/estate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "length": 100,
    "width": 100
  }'

# Getting user profile (protected)
curl -X GET http://localhost:8080/users/123-d45d-45d1-b4c5-e9bd7e86bdd1 \
  -H "Authorization: Bearer <token>"
```

### 3. Local Development
```bash
# Start with Docker Compose
export JWT_SECRET="my-dev-secret-key"
docker compose up --build

# Or run locally
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
export JWT_SECRET="my-dev-secret-key"
go run ./cmd
```

### 4. Production Deployment
Set the `JWT_SECRET` environment variable to a strong, random secret:
```bash
# Generate a strong secret (example using OpenSSL)
openssl rand -base64 32

# Set in production environment
export JWT_SECRET="your-generated-strong-secret-key"
```

---

## Token Validation Flow

```
Client Request
     ↓
  ┌─────────────────────────────────────────┐
  │ Check Authorization Header Format       │
  │ "Bearer <token>" required               │
  └─────────────────────────────────────────┘
     ↓ Valid format
  ┌─────────────────────────────────────────┐
  │ Verify JWT Signature                    │
  │ Using HMAC-SHA256 and JWT_SECRET        │
  └─────────────────────────────────────────┘
     ↓ Valid signature
  ┌─────────────────────────────────────────┐
  │ Check Token Expiration                  │
  │ Must not be older than 24 hours         │
  └─────────────────────────────────────────┘
     ↓ Not expired
  ┌─────────────────────────────────────────┐
  │ Extract User ID from Token              │
  │ Store in context for handler use        │
  └─────────────────────────────────────────┘
     ↓
  ✅ Request proceeds to protected endpoint
```

---

## Error Handling

### Missing Token
```json
{
  "Message": "Missing authorization header"
}
```

### Invalid Format
```json
{
  "Message": "Invalid authorization header format"
}
```

### Invalid/Expired Token
```json
{
  "Message": "Invalid or expired token: token has expired"
}
```

---

## Testing with Postman

1. **Login Request**
   - POST `http://localhost:8080/login`
   - Body: `{"username": "testuser", "password": "password123"}`
   - Copy the `Token` value from response

2. **Store Token**
   - Get Token → Click "..." → Postman → Set as variable
   - Or manually copy and paste into Authorization header

3. **Use Token**
   - Set Authorization type to "Bearer Token"
   - Paste token in the Token field
   - Or manually set header: `Authorization: Bearer <token>`

---

## Security Checklist

- [ ] JWT_SECRET set to strong random value in production
- [ ] Using HTTPS (not HTTP) in production
- [ ] Token expiration set appropriately (currently 24 hours)
- [ ] Validate all token errors in logs
- [ ] Rotate JWT_SECRET periodically
- [ ] Never commit JWT_SECRET to version control
- [ ] Use environment variables for secret management

---

## Common Issues

### Issue: "Invalid or expired token"
**Solution**: Generate a new token by logging in again. Tokens expire after 24 hours.

### Issue: "Invalid authorization header format"
**Solution**: Ensure header format is exactly `Authorization: Bearer <token>` (with space between Bearer and token)

### Issue: "Missing authorization header"
**Solution**: Add Authorization header for protected endpoints. Public endpoints (/login, /users POST, /hello) don't require it.

### Issue: "Failed to sign token" error in logs
**Solution**: Verify JWT_SECRET environment variable is set and not empty

---

Generated: April 13, 2026

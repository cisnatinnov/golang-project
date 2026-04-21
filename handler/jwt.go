package handler

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

const (
	// TokenExpirationTime is the duration for which JWT tokens are valid
	TokenExpirationTime = 24 * time.Hour
	// MinSecretKeyLength is the minimum length for a secure JWT secret
	MinSecretKeyLength = 32
)

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for the user
func GenerateToken(userID string, secret string) (string, error) {
	if len(secret) < MinSecretKeyLength {
		return "", fmt.Errorf("secret key must be at least %d characters", MinSecretKeyLength)
	}

	issuedAt := time.Now()
	expiresAt := issuedAt.Add(TokenExpirationTime)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   userID,
			Issuer:    "UserService",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the user ID
func ValidateToken(tokenString string, secret string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims.UserID == "" {
		return "", fmt.Errorf("invalid token: missing user_id")
	}

	return claims.UserID, nil
}

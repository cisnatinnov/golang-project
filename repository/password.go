package repository

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing
	// Higher values increase security but also computation time
	BcryptCost = 12
)

// HashPassword hashes a plaintext password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", &ValidationError{Message: "password must be at least 8 characters"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a plaintext password matches a hashed password
func VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

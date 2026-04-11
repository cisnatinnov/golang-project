package handler

import (
	"net/http"
	"strings"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/labstack/echo/v4"
)

const (
	// ContextKeyUserID is the key for storing user ID in context
	ContextKeyUserID = "user_id"
)

// BearerTokenMiddleware validates bearer token in Authorization header
func (s *Server) BearerTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, generated.ErrorResponse{
				Message: "Missing authorization header",
			})
		}

		// Extract bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, generated.ErrorResponse{
				Message: "Invalid authorization header format",
			})
		}

		token := parts[1]
		if token == "" {
			return c.JSON(http.StatusUnauthorized, generated.ErrorResponse{
				Message: "Invalid bearer token",
			})
		}

		// Extract user ID from token
		// Token format: placeholder_token_{user_id}
		userID := extractUserIDFromToken(token)
		if userID == "" {
			return c.JSON(http.StatusUnauthorized, generated.ErrorResponse{
				Message: "Invalid token format",
			})
		}

		// Store user ID in context
		c.Set(ContextKeyUserID, userID)

		return next(c)
	}
}

// extractUserIDFromToken extracts user ID from token
// Token format: placeholder_token_{user_id}
func extractUserIDFromToken(token string) string {
	prefix := "placeholder_token_"
	if !strings.HasPrefix(token, prefix) {
		return ""
	}
	userID := strings.TrimPrefix(token, prefix)
	if userID == "" {
		return ""
	}
	return userID
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c echo.Context) string {
	userID := c.Get(ContextKeyUserID)
	if userID == nil {
		return ""
	}
	return userID.(string)
}

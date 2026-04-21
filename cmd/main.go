package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/handler"
	"github.com/SawitProRecruitment/UserService/repository"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// DefaultJWTSecret is used only for development when JWT_SECRET is not set
	DefaultJWTSecret = "your-secret-key-change-in-production-must-be-32-chars"
	// ServerAddress is the default address the server listens on
	ServerAddress = ":1323"
	// ShutdownTimeout is the timeout for graceful shutdown
	ShutdownTimeout = 10 * time.Second
)

func main() {
	e := echo.New()

	// Configure Echo with security middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	e.Use(middleware.RequestID())

	s := newServer()
	var server generated.ServerInterface = s

	// Apply bearer token middleware for protected routes
	e.Use(s.BearerTokenMiddlewareWithSkipper())
	generated.RegisterHandlers(e, server)

	// Start server in a goroutine
	go func() {
		if err := e.Start(ServerAddress); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	fmt.Printf("Server started on %s\n", ServerAddress)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	// Close repository database connection
	if repo, ok := s.Repository.(*repository.Repository); ok {
		if err := repo.Close(); err != nil {
			fmt.Printf("Error closing database connection: %v\n", err)
		}
	}

	fmt.Println("Server exited")
}

func newServer() *handler.Server {
	dbDsn := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Use a default secret if not provided (for development only)
	if jwtSecret == "" {
		fmt.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production.")
		jwtSecret = DefaultJWTSecret
	}

	// Validate JWT secret length
	if len(jwtSecret) < handler.MinSecretKeyLength {
		panic(fmt.Sprintf("JWT_SECRET must be at least %d characters", handler.MinSecretKeyLength))
	}

	var repo repository.RepositoryInterface = repository.NewRepository(repository.NewRepositoryOptions{
		Dsn: dbDsn,
	})
	opts := handler.NewServerOptions{
		Repository: repo,
		JWTSecret:  jwtSecret,
	}
	return handler.NewServer(opts)
}

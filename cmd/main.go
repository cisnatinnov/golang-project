package main

import (
	"os"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/handler"
	"github.com/SawitProRecruitment/UserService/repository"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()

	s := newServer()
	var server generated.ServerInterface = s

	// Apply bearer token middleware for protected routes
	e.Use(s.BearerTokenMiddlewareWithSkipper())
	generated.RegisterHandlers(e, server)
	e.Use(middleware.Logger())
	e.Logger.Fatal(e.Start(":1323"))
}

func newServer() *handler.Server {
	dbDsn := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Use a default secret if not provided (for development only)
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
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

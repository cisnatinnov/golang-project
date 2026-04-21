package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	// DefaultConnectionTimeout is the default timeout for database connections
	DefaultConnectionTimeout = 5 * time.Second
	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns = 25
	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns = 5
	// ConnMaxLifetime is the maximum amount of time a connection may be reused
	ConnMaxLifetime = 5 * time.Minute
)

type Repository struct {
	Db *sql.DB
}

type NewRepositoryOptions struct {
	Dsn string
}

func NewRepository(opts NewRepositoryOptions) *Repository {
	if opts.Dsn == "" {
		panic("database DSN is required")
	}

	db, err := sql.Open("postgres", opts.Dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to open database connection: %v", err))
	}

	// Configure connection pool
	db.SetMaxOpenConns(MaxOpenConns)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetConnMaxLifetime(ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectionTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	return &Repository{
		Db: db,
	}
}

// Close closes the database connection
func (r *Repository) Close() error {
	if r.Db != nil {
		return r.Db.Close()
	}
	return nil
}

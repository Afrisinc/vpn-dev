package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool   *pgxpool.Pool
	once   sync.Once
	dbOnce sync.Once
)

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewPostgresConnection creates a new PostgreSQL connection pool
func NewPostgresConnection(cfg PostgresConfig) (*pgxpool.Pool, error) {
	var err error
	once.Do(func() {
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			cfg.SSLMode,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return
		}

		// Test the connection
		err = pool.Ping(ctx)
		if err != nil {
			pool = nil
		}
	})
	return pool, err
}

// GetPool returns the singleton PostgreSQL connection pool
func GetPool() *pgxpool.Pool {
	return pool
}

// GetDB returns a database connection from the pool
func GetDB(ctx context.Context) (*pgxpool.Conn, error) {
	if pool == nil {
		return nil, fmt.Errorf("database pool not initialized")
	}
	return pool.Acquire(ctx)
}

// ClosePool closes the database connection pool
func ClosePool() {
	if pool != nil {
		pool.Close()
	}
}

// PingDB tests the database connection
func PingDB(ctx context.Context) error {
	if pool == nil {
		return fmt.Errorf("database pool not initialized")
	}
	return pool.Ping(ctx)
}

// GetDatabase returns the connection for use in repositories
// This is kept simple for now; can be extended for transaction support
func GetDatabase() *pgxpool.Pool {
	return pool
}

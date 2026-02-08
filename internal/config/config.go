package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Database      DatabaseConfig
	Server        ServerConfig
	Security      SecurityConfig
	WireGuard     WireGuardConfig
	IsDevelopment bool
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Env  string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	RateLimitRequests int
	RateLimitWindow   string
	MaxRequestSize    int64
}

// WireGuardConfig holds WireGuard configuration
type WireGuardConfig struct {
	InterfaceName    string
	ServerPublicKey  string
	ServerPrivateKey string
	ServerEndpoint   string
	ListenPort       string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file (optional, won't fail if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "vpn_user"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			Database: getEnv("POSTGRES_DB", "vpn_dev"),
			SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Security: SecurityConfig{
			RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 10),
			RateLimitWindow:   getEnv("RATE_LIMIT_WINDOW", "1m"),
			MaxRequestSize:    getEnvInt64("MAX_REQUEST_SIZE", 1048576), // 1MB
		},
		WireGuard: WireGuardConfig{
			InterfaceName:    getEnv("WG_INTERFACE_NAME", "wg0"),
			ServerPublicKey:  getEnv("WG_SERVER_PUBLIC_KEY", "cWC4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y="),
			ServerPrivateKey: getEnv("WG_SERVER_PRIVATE_KEY", "uHXTbxAQJ3kO2gXwGui3p+/F1XYhYccXe1jBW/3wEnI="),
			ServerEndpoint:   getEnv("WG_SERVER_ENDPOINT", "192.168.1.71:51820"),
			ListenPort:       getEnv("WG_LISTEN_PORT", "51820"),
		},
		IsDevelopment: getEnv("SERVER_ENV", "development") == "development",
	}

	return cfg
}

// ConnectDatabase establishes a PostgreSQL connection pool
func ConnectDatabase(cfg *Config) (*pgxpool.Pool, error) {
	dbCfg := db.PostgresConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
	}

	pool, err := db.NewPostgresConnection(dbCfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL")
	return pool, nil
}

// getEnv retrieves an environment variable with a fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt retrieves an integer environment variable with a fallback
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvInt64 retrieves an int64 environment variable with a fallback
func getEnvInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetDurationParsed parses duration string (e.g., "1m", "30s")
func GetDurationParsed(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Minute
	}
	return d
}

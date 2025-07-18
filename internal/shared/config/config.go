package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application
type Config struct {
	HTTPAddr    string `envconfig:"HTTP_ADDR" default:":8080"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`

	// Database configuration
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"testuser"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"testpass"`
	DBName     string `envconfig:"DB_NAME" default:"testdb"`
	DBSSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`

	// Session configuration
	SessionSecret string `envconfig:"SESSION_SECRET" default:"your-secret-key-change-in-production"`

	// Admin bootstrap configuration
	AdminEmail    string `envconfig:"ADMIN_EMAIL" default:"admin@test.local"`
	AdminPassword string `envconfig:"ADMIN_PASSWORD" default:"admin123"`
	AdminName     string `envconfig:"ADMIN_NAME" default:"Administrator"`
}

// Parse loads configuration from environment variables
func (c *Config) Parse() error {
	return envconfig.Process("", c)
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	host := c.DBHost
	// Use explicit IPv4 address instead of localhost to avoid IPv6 issues
	if host == "localhost" {
		host = "127.0.0.1"
	}
	return "host=" + host + " port=" + c.DBPort + " user=" + c.DBUser +
		" password=" + c.DBPassword + " dbname=" + c.DBName + " sslmode=" + c.DBSSLMode
}

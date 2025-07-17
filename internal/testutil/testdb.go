package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/acheevo/test/internal/repository"
)

// TestDB represents a test database container
type TestDB struct {
	Container *postgres.PostgresContainer
	Database  *repository.Database
	DSN       string
}

// SetupTestDB creates a test database container and returns the connection
func SetupTestDB(t *testing.T) *TestDB {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Create DSN
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host, port.Port())

	// Create database connection
	db, err := repository.NewDatabase(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Enable uuid-ossp extension for PostgreSQL
	if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		t.Logf("Warning: Failed to create uuid-ossp extension: %v", err)
	}

	return &TestDB{
		Container: postgresContainer,
		Database:  db,
		DSN:       dsn,
	}
}

// Cleanup closes the database connection and terminates the container
func (tdb *TestDB) Cleanup(t *testing.T) {
	ctx := context.Background()

	if tdb.Database != nil {
		if err := tdb.Database.Close(); err != nil {
			t.Errorf("Failed to close database connection: %v", err)
		}
	}

	if tdb.Container != nil {
		if err := tdb.Container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}
}
package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	authDomain "github.com/acheevo/test/internal/auth/domain"
	userDomain "github.com/acheevo/test/internal/user/domain"
)

// Database represents the database connection and operations
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Enable uuid-ossp extension for PostgreSQL
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		// Log warning but don't fail - some databases might not support this
		// or might have it enabled already
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&userDomain.User{}, &authDomain.Session{}); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/acheevo/test/internal/auth/domain"
	"github.com/acheevo/test/internal/shared/database"
)

// SessionRepository handles session-related database operations
type SessionRepository struct {
	db *database.Database
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *database.Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(session *domain.Session) error {
	return r.db.DB.Create(session).Error
}

// GetByToken retrieves a session by token
func (r *SessionRepository) GetByToken(token string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, err
}

// DeleteByToken deletes a session by token
func (r *SessionRepository) DeleteByToken(token string) error {
	return r.db.DB.Where("token = ?", token).Delete(&domain.Session{}).Error
}

// DeleteExpired deletes expired sessions
func (r *SessionRepository) DeleteExpired() error {
	return r.db.DB.Where("expires_at < ?", time.Now()).Delete(&domain.Session{}).Error
}

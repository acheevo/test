package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/acheevo/test/internal/models"
)

// SessionRepository handles session-related database operations
type SessionRepository struct {
	db *Database
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(session *models.Session) error {
	return r.db.DB.Create(session).Error
}

// GetByToken retrieves a session by token
func (r *SessionRepository) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, err
}

// DeleteByToken deletes a session by token
func (r *SessionRepository) DeleteByToken(token string) error {
	return r.db.DB.Where("token = ?", token).Delete(&models.Session{}).Error
}

// DeleteExpired deletes expired sessions
func (r *SessionRepository) DeleteExpired() error {
	return r.db.DB.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}
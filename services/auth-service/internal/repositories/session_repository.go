package repositories

import (
	"time"

	"velure-auth-service/internal/models"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) GetByUserID(userID uint) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("user_id = ?", userID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetByRefreshToken(refreshToken string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("refresh_token = ?", refreshToken).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Update(session *models.Session) error {
	return r.db.Save(session).Error
}

func (r *SessionRepository) InvalidateByRefreshToken(refreshToken string) error {
	return r.db.Model(&models.Session{}).
		Where("refresh_token = ?", refreshToken).
		Update("expires_at", time.Now()).Error
}

func (r *SessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Session{}, id).Error
}

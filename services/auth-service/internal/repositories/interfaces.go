package repositories

import "velure-auth-service/internal/models"

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetAll() ([]models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	GetByPage(page, pageSize int) ([]models.User, int64, error)
}

// SessionRepositoryInterface defines the interface for session repository operations
type SessionRepositoryInterface interface {
	Create(session *models.Session) error
	GetByUserID(userID uint) (*models.Session, error)
	GetByRefreshToken(refreshToken string) (*models.Session, error)
	Update(session *models.Session) error
	InvalidateByRefreshToken(refreshToken string) error
	Delete(id uint) error
}

// PasswordResetRepositoryInterface defines the interface for password reset repository operations
type PasswordResetRepositoryInterface interface {
	Create(passwordReset *models.PasswordReset) error
	GetByToken(token string) (*models.PasswordReset, error)
	Delete(id uint) error
}

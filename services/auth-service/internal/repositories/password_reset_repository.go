package repositories

import (
	"velure-auth-service/internal/models"

	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(passwordReset *models.PasswordReset) error {
	return r.db.Create(passwordReset).Error
}

func (r *PasswordResetRepository) GetByToken(token string) (*models.PasswordReset, error) {
	var passwordReset models.PasswordReset
	err := r.db.Where("token = ?", token).First(&passwordReset).Error
	if err != nil {
		return nil, err
	}
	return &passwordReset, nil
}

func (r *PasswordResetRepository) Delete(id uint) error {
	return r.db.Delete(&models.PasswordReset{}, id).Error
}

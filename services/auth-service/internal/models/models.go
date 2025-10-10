package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Sessions       []Session       `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	PasswordResets []PasswordReset `json:"passwordResets,omitempty" gorm:"foreignKey:UserID"`
}

type Session struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"not null"`
	AccessToken  string    `json:"accessToken" gorm:"unique;not null"`
	RefreshToken string    `json:"refreshToken" gorm:"unique;not null"`
	ExpiresAt    time.Time `json:"expiresAt" gorm:"not null"`
	CreatedAt    time.Time `json:"createdAt"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type PasswordReset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// DTOs
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ValidateTokenRequest struct {
	AccessToken string `json:"token" binding:"required"`
}

type ValidateTokenResponse struct {
	IsValid bool `json:"isValid"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

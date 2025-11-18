package services

import "velure-auth-service/internal/models"

// AuthServiceInterface defines the interface for authentication service operations
type AuthServiceInterface interface {
	CreateUser(req models.CreateUserRequest) (*models.RegistrationResponse, error)
	Login(req models.LoginRequest) (*models.LoginResponse, error)
	ValidateAccessToken(token string) (*models.User, error)
	GetUsers() ([]models.UserResponse, error)
	GetUsersByPage(page, pageSize int) (*models.PaginatedUsersResponse, error)
	GetUserByID(id uint) (*models.UserResponse, error)
	GetUserByEmail(email string) (*models.UserResponse, error)
	Logout(refreshToken string) error
}

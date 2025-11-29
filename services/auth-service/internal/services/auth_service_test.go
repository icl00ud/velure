package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"velure-auth-service/internal/metrics"
	"velure-auth-service/internal/mocks"
	"velure-auth-service/internal/models"
	"velure-auth-service/internal/testutil"

	"github.com/go-redis/redismock/v9"
	"github.com/golang-jwt/jwt/v5"
	promtest "github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		req       models.CreateUserRequest
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful user creation",
			req: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByEmail("test@example.com").
					Return(nil, gorm.ErrRecordNotFound)

				mockUserRepo.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(u *models.User) error {
						u.ID = 1
						u.CreatedAt = time.Now()
						u.UpdatedAt = time.Now()
						return nil
					})

				mockSessionRepo.EXPECT().
					GetByUserID(uint(1)).
					Return(nil, gorm.ErrRecordNotFound)

				mockSessionRepo.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(s *models.Session) error {
						s.ID = 1
						s.AccessToken = "test-access-token"
						s.RefreshToken = "test-refresh-token"
						s.ExpiresAt = time.Now().Add(24 * time.Hour)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "duplicate email error",
			req: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface) {
				existingUser := &models.User{ID: 1, Email: "existing@example.com"}
				mockUserRepo.EXPECT().
					GetByEmail("existing@example.com").
					Return(existingUser, nil)
			},
			wantErr: true,
			errMsg:  "user already exists",
		},
		{
			name: "database error checking existing user",
			req: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByEmail("test@example.com").
					Return(nil, errors.New("database connection error"))
			},
			wantErr: true,
			errMsg:  "error checking existing user",
		},
		{
			name: "database error on create",
			req: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByEmail("test@example.com").
					Return(nil, gorm.ErrRecordNotFound)

				mockUserRepo.EXPECT().
					Create(gomock.Any()).
					Return(errors.New("database insert error"))
			},
			wantErr: true,
			errMsg:  "error creating user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo, mockSessionRepo)

			result, err := service.CreateUser(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateUser() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("CreateUser() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CreateUser() unexpected error = %v", err)
				}
				if result == nil {
					t.Error("CreateUser() expected result but got nil")
				} else {
					if result.Email != tt.req.Email {
						t.Errorf("CreateUser() email = %v, want %v", result.Email, tt.req.Email)
					}
					if result.Name != tt.req.Name {
						t.Errorf("CreateUser() name = %v, want %v", result.Name, tt.req.Name)
					}
					if result.AccessToken == "" {
						t.Error("CreateUser() expected accessToken but got empty string")
					}
					if result.RefreshToken == "" {
						t.Error("CreateUser() expected refreshToken but got empty string")
					}
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		req       models.LoginRequest
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface, hashedPwd string)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful login",
			req: models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface, hashedPwd string) {
				user := &models.User{
					ID:       1,
					Email:    "user@example.com",
					Password: hashedPwd,
				}
				mockUserRepo.EXPECT().
					GetByEmail("user@example.com").
					Return(user, nil)

				mockSessionRepo.EXPECT().
					GetByUserID(uint(1)).
					Return(nil, gorm.ErrRecordNotFound)

				mockSessionRepo.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(s *models.Session) error {
						s.ID = 1
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "user not found",
			req: models.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface, hashedPwd string) {
				mockUserRepo.EXPECT().
					GetByEmail("nonexistent@example.com").
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errMsg:  "invalid credentials",
		},
		{
			name: "invalid password",
			req: models.LoginRequest{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface, hashedPwd string) {
				user := &models.User{
					ID:       1,
					Email:    "user@example.com",
					Password: hashedPwd,
				}
				mockUserRepo.EXPECT().
					GetByEmail("user@example.com").
					Return(user, nil)
			},
			wantErr: true,
			errMsg:  "invalid credentials",
		},
		{
			name: "update existing session",
			req: models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface, hashedPwd string) {
				user := &models.User{
					ID:       1,
					Email:    "user@example.com",
					Password: hashedPwd,
				}
				mockUserRepo.EXPECT().
					GetByEmail("user@example.com").
					Return(user, nil)

				existingSession := &models.Session{
					ID:     1,
					UserID: 1,
				}
				mockSessionRepo.EXPECT().
					GetByUserID(uint(1)).
					Return(existingSession, nil)

				mockSessionRepo.EXPECT().
					Update(gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error on get user",
			req: models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface, mockSessionRepo *mocks.MockSessionRepositoryInterface, hashedPwd string) {
				mockUserRepo.EXPECT().
					GetByEmail("user@example.com").
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "error getting user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo, mockSessionRepo, string(hashedPassword))

			result, err := service.Login(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Login() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Login() unexpected error = %v", err)
				}
				if result == nil {
					t.Error("Login() expected result but got nil")
				} else {
					if result.AccessToken == "" {
						t.Error("Login() AccessToken should not be empty")
					}
					if result.RefreshToken == "" {
						t.Error("Login() RefreshToken should not be empty")
					}
				}
			}
		})
	}
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	cfg := testutil.CreateTestConfig()
	validToken := generateTestToken(t, cfg.JWT.Secret, 1, time.Now().Add(1*time.Hour))
	expiredToken := generateTestToken(t, cfg.JWT.Secret, 1, time.Now().Add(-1*time.Hour))

	tests := []struct {
		name      string
		token     string
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface)
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "valid token",
			token: validToken,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				user := &models.User{
					ID:    1,
					Email: "user@example.com",
					Name:  "Test User",
				}
				mockUserRepo.EXPECT().
					GetByID(uint(1)).
					Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:      "expired token",
			token:     expiredToken,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {},
			wantErr:   true,
			errMsg:    "invalid token",
		},
		{
			name:      "invalid token format",
			token:     "invalid-token-string",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {},
			wantErr:   true,
			errMsg:    "invalid token",
		},
		{
			name:  "user not found",
			token: validToken,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByID(uint(1)).
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:  "database error",
			token: validToken,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByID(uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo)

			result, err := service.ValidateAccessToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAccessToken() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateAccessToken() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAccessToken() unexpected error = %v", err)
				}
				if result == nil {
					t.Error("ValidateAccessToken() expected user but got nil")
				} else if result.ID != 1 {
					t.Errorf("ValidateAccessToken() user ID = %v, want 1", result.ID)
				}
			}
		})
	}
}

func TestAuthService_GetUsers(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface)
		wantErr   bool
		wantCount int
	}{
		{
			name: "successful get all users",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				users := []models.User{
					{ID: 1, Email: "user1@example.com", Name: "User 1"},
					{ID: 2, Email: "user2@example.com", Name: "User 2"},
				}
				mockUserRepo.EXPECT().
					GetAll().
					Return(users, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "empty user list",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetAll().
					Return([]models.User{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "database error",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetAll().
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo)

			result, err := service.GetUsers()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetUsers() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetUsers() unexpected error = %v", err)
				}
				if len(result) != tt.wantCount {
					t.Errorf("GetUsers() returned %d users, want %d", len(result), tt.wantCount)
				}
			}
		})
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "user found",
			userID: 1,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				user := &models.User{
					ID:    1,
					Email: "user@example.com",
					Name:  "Test User",
				}
				mockUserRepo.EXPECT().
					GetByID(uint(1)).
					Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByID(uint(999)).
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:   "database error",
			userID: 1,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByID(uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "error getting user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo)

			result, err := service.GetUserByID(tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetUserByID() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetUserByID() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("GetUserByID() unexpected error = %v", err)
				}
				if result == nil {
					t.Error("GetUserByID() expected result but got nil")
				} else if result.ID != tt.userID {
					t.Errorf("GetUserByID() ID = %v, want %v", result.ID, tt.userID)
				}
			}
		})
	}
}

func TestAuthService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface)
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "user found",
			email: "user@example.com",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				user := &models.User{
					ID:    1,
					Email: "user@example.com",
					Name:  "Test User",
				}
				mockUserRepo.EXPECT().
					GetByEmail("user@example.com").
					Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByEmail("nonexistent@example.com").
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:  "database error",
			email: "user@example.com",
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByEmail("user@example.com").
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "error getting user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo)

			result, err := service.GetUserByEmail(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetUserByEmail() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetUserByEmail() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("GetUserByEmail() unexpected error = %v", err)
				}
				if result == nil {
					t.Error("GetUserByEmail() expected result but got nil")
				} else if result.Email != tt.email {
					t.Errorf("GetUserByEmail() email = %v, want %v", result.Email, tt.email)
				}
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		setupMock    func(mockSessionRepo *mocks.MockSessionRepositoryInterface)
		wantErr      bool
	}{
		{
			name:         "successful logout",
			refreshToken: "valid-refresh-token",
			setupMock: func(mockSessionRepo *mocks.MockSessionRepositoryInterface) {
				mockSessionRepo.EXPECT().
					InvalidateByRefreshToken("valid-refresh-token").
					Return(nil)
				// Logout calls SyncActiveSessionsMetric which calls CountActiveSessions
				mockSessionRepo.EXPECT().
					CountActiveSessions(gomock.Any()).
					Return(int64(0), nil).
					AnyTimes()
			},
			wantErr: false,
		},
		{
			name:         "database error",
			refreshToken: "token",
			setupMock: func(mockSessionRepo *mocks.MockSessionRepositoryInterface) {
				mockSessionRepo.EXPECT().
					InvalidateByRefreshToken("token").
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockSessionRepo)

			err := service.Logout(tt.refreshToken)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Logout() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Logout() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestAuthService_GetUsersByPage(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		pageSize  int
		setupMock func(mockUserRepo *mocks.MockUserRepositoryInterface)
		wantErr   bool
		wantTotal int64
		wantCount int
	}{
		{
			name:     "successful pagination",
			page:     1,
			pageSize: 10,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				users := []models.User{
					{ID: 1, Email: "user1@example.com", Name: "User 1"},
					{ID: 2, Email: "user2@example.com", Name: "User 2"},
				}
				mockUserRepo.EXPECT().
					GetByPage(1, 10).
					Return(users, int64(2), nil)
			},
			wantErr:   false,
			wantTotal: 2,
			wantCount: 2,
		},
		{
			name:     "database error",
			page:     1,
			pageSize: 10,
			setupMock: func(mockUserRepo *mocks.MockUserRepositoryInterface) {
				mockUserRepo.EXPECT().
					GetByPage(1, 10).
					Return(nil, int64(0), errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
			mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
			cfg := testutil.CreateTestConfig()

			service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

			tt.setupMock(mockUserRepo)

			result, err := service.GetUsersByPage(tt.page, tt.pageSize)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetUsersByPage() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetUsersByPage() unexpected error = %v", err)
				}
				if result == nil {
					t.Error("GetUsersByPage() expected result but got nil")
				} else {
					if result.TotalCount != tt.wantTotal {
						t.Errorf("GetUsersByPage() total = %v, want %v", result.TotalCount, tt.wantTotal)
					}
					if len(result.Users) != tt.wantCount {
						t.Errorf("GetUsersByPage() count = %v, want %v", len(result.Users), tt.wantCount)
					}
				}
			}
		})
	}
}

func TestAuthService_UpdateOrCreateSessionAsync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)

	mockSessionRepo.EXPECT().
		GetByUserID(uint(1)).
		Return(nil, gorm.ErrRecordNotFound)

	mockSessionRepo.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(s *models.Session) error {
			s.ID = 99
			return nil
		})

	session, err := service.updateOrCreateSessionAsync(1)
	if err != nil {
		t.Fatalf("updateOrCreateSessionAsync() error = %v", err)
	}
	if session.ID != 99 {
		t.Fatalf("expected session ID to be set, got %d", session.ID)
	}
}

func TestAuthService_SyncTotalUsersMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	metrics.TotalUsers.Set(0)

	mockUserRepo.EXPECT().
		CountUsers(gomock.Any()).
		Return(int64(7), nil)

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)
	service.SyncTotalUsersMetric(context.Background())

	if got := promtest.ToFloat64(metrics.TotalUsers); got != 7 {
		t.Fatalf("expected total users metric to be 7, got %.0f", got)
	}
}

func TestAuthService_SyncTotalUsersMetric_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	metrics.TotalUsers.Set(3)

	mockUserRepo.EXPECT().
		CountUsers(gomock.Any()).
		Return(int64(0), errors.New("db error"))

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)
	service.SyncTotalUsersMetric(context.Background())

	if got := promtest.ToFloat64(metrics.TotalUsers); got != 3 {
		t.Fatalf("expected metric to remain unchanged on error, got %.0f", got)
	}
}

func TestAuthService_ValidateAccessToken_UsesCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()
	cfg.Performance.EnableCache = true
	cfg.Performance.TokenCacheTTL = 1

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)
	token := generateTestToken(t, cfg.JWT.Secret, 1, time.Now().Add(time.Hour))

	user := &models.User{ID: 1, Email: "cached@example.com"}
	mockUserRepo.EXPECT().
		GetByID(uint(1)).
		Return(user, nil)

	firstUser, err := service.ValidateAccessToken(token)
	if err != nil || firstUser.ID != 1 {
		t.Fatalf("first ValidateAccessToken() call failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	secondUser, err := service.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("expected cached validation to succeed, got %v", err)
	}
	if secondUser.Email != user.Email {
		t.Fatalf("unexpected cached user: %#v", secondUser)
	}
}

func TestAuthService_GetUserByID_FromRedisCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	user := &models.User{ID: 1, Email: "cached@example.com", Name: "Cached"}
	data, _ := json.Marshal(user)

	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectGet("user:id:1").SetVal(string(data))

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, redisClient)

	result, err := service.GetUserByID(1)
	if err != nil {
		t.Fatalf("GetUserByID() unexpected error: %v", err)
	}
	if result.Email != user.Email {
		t.Fatalf("expected cached user email %s, got %s", user.Email, result.Email)
	}

	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestAuthService_GetUserByEmail_FromRedisCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	user := &models.User{ID: 2, Email: "cache@test.com", Name: "Cache User"}
	data, _ := json.Marshal(user)

	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectGet("user:email:cache@test.com").SetVal(string(data))

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, redisClient)

	result, err := service.GetUserByEmail("cache@test.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() unexpected error: %v", err)
	}
	if result.ID != user.ID {
		t.Fatalf("expected cached user id %d, got %d", user.ID, result.ID)
	}

	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestAuthService_GetUserByID_CacheMissWithRedis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectGet("user:id:3").RedisNil()

	mockUserRepo.EXPECT().
		GetByID(uint(3)).
		Return(&models.User{ID: 3, Email: "miss@example.com"}, nil)

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, redisClient)

	_, err := service.GetUserByID(3)
	if err != nil {
		t.Fatalf("expected successful fetch after cache miss, got %v", err)
	}

	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestAuthService_SyncActiveSessionsMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	metrics.ActiveSessions.Set(0)

	mockSessionRepo.EXPECT().
		CountActiveSessions(gomock.Any()).
		Return(int64(4), nil)

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)
	service.SyncActiveSessionsMetric(nil)

	if got := promtest.ToFloat64(metrics.ActiveSessions); got != 4 {
		t.Fatalf("expected active sessions metric to be 4, got %.0f", got)
	}
}

func TestAuthService_SyncActiveSessionsMetric_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)
	cfg := testutil.CreateTestConfig()

	metrics.ActiveSessions.Set(2)

	mockSessionRepo.EXPECT().
		CountActiveSessions(gomock.Any()).
		Return(int64(0), errors.New("count error"))

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg, nil)
	service.SyncActiveSessionsMetric(nil)

	if got := promtest.ToFloat64(metrics.ActiveSessions); got != 2 {
		t.Fatalf("expected metric to remain unchanged on error, got %.0f", got)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func generateTestToken(t *testing.T, secret string, userID uint, expiresAt time.Time) string {
	claims := jwt.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return tokenString
}

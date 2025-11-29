package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"velure-auth-service/internal/mocks"
	"velure-auth-service/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful registration",
			requestBody: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockService.EXPECT().
					CreateUser(gomock.Any()).
					Return(&models.RegistrationResponse{
						ID:           1,
						Name:         "Test User",
						Email:        "test@example.com",
						AccessToken:  "test-access-token",
						RefreshToken: "test-refresh-token",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["email"] != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %v", body["email"])
				}
				if body["accessToken"] != "test-access-token" {
					t.Errorf("Expected accessToken test-access-token, got %v", body["accessToken"])
				}
				if body["refreshToken"] != "test-refresh-token" {
					t.Errorf("Expected refreshToken test-refresh-token, got %v", body["refreshToken"])
				}
			},
		},
		{
			name: "duplicate email",
			requestBody: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockService.EXPECT().
					CreateUser(gomock.Any()).
					Return(nil, errors.New("user already exists"))
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] != "user already exists" {
					t.Errorf("Expected error 'user already exists', got %v", body["error"])
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name: "internal server error",
			requestBody: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockService.EXPECT().
					CreateUser(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.POST("/register", handler.Register)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			tt.checkResponse(t, responseBody)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockService.EXPECT().
					Login(gomock.Any()).
					Return(&models.LoginResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["accessToken"] != "access-token" {
					t.Errorf("Expected accessToken, got %v", body["accessToken"])
				}
				if body["refreshToken"] != "refresh-token" {
					t.Errorf("Expected refreshToken, got %v", body["refreshToken"])
				}
			},
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			setupMock: func() {
				mockService.EXPECT().
					Login(gomock.Any()).
					Return(nil, errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] != "invalid credentials" {
					t.Errorf("Expected error 'invalid credentials', got %v", body["error"])
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name: "internal server error",
			requestBody: models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockService.EXPECT().
					Login(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.POST("/login", handler.Login)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			tt.checkResponse(t, responseBody)
		})
	}
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "valid token",
			requestBody: models.ValidateTokenRequest{
				AccessToken: "valid-token",
			},
			setupMock: func() {
				mockService.EXPECT().
					ValidateAccessToken("valid-token").
					Return(&models.User{ID: 1, Email: "user@example.com"}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["isValid"] != true {
					t.Errorf("Expected isValid=true, got %v", body["isValid"])
				}
			},
		},
		{
			name: "invalid token",
			requestBody: models.ValidateTokenRequest{
				AccessToken: "invalid-token",
			},
			setupMock: func() {
				mockService.EXPECT().
					ValidateAccessToken("invalid-token").
					Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["isValid"] != false {
					t.Errorf("Expected isValid=false, got %v", body["isValid"])
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.POST("/validate", handler.ValidateToken)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			tt.checkResponse(t, responseBody)
		})
	}
}

func TestAuthHandler_GetUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:        "get all users",
			queryParams: "",
			setupMock: func() {
				mockService.EXPECT().
					GetUsers().
					Return([]models.UserResponse{
						{ID: 1, Name: "User 1", Email: "user1@example.com"},
						{ID: 2, Name: "User 2", Email: "user2@example.com"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var users []models.UserResponse
				json.Unmarshal(w.Body.Bytes(), &users)
				if len(users) != 2 {
					t.Errorf("Expected 2 users, got %d", len(users))
				}
			},
		},
		{
			name:        "get users with pagination",
			queryParams: "?page=1&pageSize=10",
			setupMock: func() {
				mockService.EXPECT().
					GetUsersByPage(1, 10).
					Return(&models.PaginatedUsersResponse{
						Users:      []models.UserResponse{{ID: 1, Name: "User 1", Email: "user1@example.com"}},
						TotalCount: 1,
						Page:       1,
						PageSize:   10,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.PaginatedUsersResponse
				json.Unmarshal(w.Body.Bytes(), &response)
				if len(response.Users) != 1 {
					t.Errorf("Expected 1 user, got %d", len(response.Users))
				}
			},
		},
		{
			name:        "database error",
			queryParams: "",
			setupMock: func() {
				mockService.EXPECT().
					GetUsers().
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var body map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &body)
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name:        "pagination error",
			queryParams: "?page=1&pageSize=10",
			setupMock: func() {
				mockService.EXPECT().
					GetUsersByPage(1, 10).
					Return(nil, errors.New("paginate error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var body map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &body)
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.GET("/users", handler.GetUsers)

			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			tt.checkResponse(t, w)
		})
	}
}

func TestAuthHandler_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		userID         string
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:   "user found",
			userID: "1",
			setupMock: func() {
				mockService.EXPECT().
					GetUserByID(uint(1)).
					Return(&models.UserResponse{
						ID:    1,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["email"] != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %v", body["email"])
				}
			},
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] != "invalid user ID" {
					t.Errorf("Expected error 'invalid user ID', got %v", body["error"])
				}
			},
		},
		{
			name:   "user not found",
			userID: "999",
			setupMock: func() {
				mockService.EXPECT().
					GetUserByID(uint(999)).
					Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] != "user not found" {
					t.Errorf("Expected error 'user not found', got %v", body["error"])
				}
			},
		},
		{
			name:   "internal server error",
			userID: "1",
			setupMock: func() {
				mockService.EXPECT().
					GetUserByID(uint(1)).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.GET("/users/:id", handler.GetUserByID)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			tt.checkResponse(t, responseBody)
		})
	}
}

func TestAuthHandler_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		email          string
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:  "user found",
			email: "test@example.com",
			setupMock: func() {
				mockService.EXPECT().
					GetUserByEmail("test@example.com").
					Return(&models.UserResponse{
						ID:    1,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["email"] != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %v", body["email"])
				}
			},
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMock: func() {
				mockService.EXPECT().
					GetUserByEmail("nonexistent@example.com").
					Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] != "user not found" {
					t.Errorf("Expected error 'user not found', got %v", body["error"])
				}
			},
		},
		{
			name:  "internal server error",
			email: "test@example.com",
			setupMock: func() {
				mockService.EXPECT().
					GetUserByEmail("test@example.com").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.GET("/users/email/:email", handler.GetUserByEmail)

			req := httptest.NewRequest(http.MethodGet, "/users/email/"+tt.email, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			tt.checkResponse(t, responseBody)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockAuthServiceInterface(ctrl)
	handler := NewAuthHandler(mockService)

	tests := []struct {
		name           string
		refreshToken   string
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:         "successful logout",
			refreshToken: "valid-refresh-token",
			setupMock: func() {
				mockService.EXPECT().
					Logout("valid-refresh-token").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["message"] != "logout successful" {
					t.Errorf("Expected success message, got %v", body["message"])
				}
			},
		},
		{
			name:         "internal server error",
			refreshToken: "token",
			setupMock: func() {
				mockService.EXPECT().
					Logout("token").
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			router := setupTestRouter()
			router.DELETE("/logout/:refreshToken", handler.Logout)

			req := httptest.NewRequest(http.MethodDelete, "/logout/"+tt.refreshToken, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			tt.checkResponse(t, responseBody)
		})
	}
}

func TestAuthHandler_GetUserByEmail_MissingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "email", Value: ""}}

	handler.GetUserByEmail(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "refreshToken", Value: ""}}

	handler.Logout(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

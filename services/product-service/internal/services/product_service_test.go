package services

import (
	"context"
	"errors"
	"testing"

	"product-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductResponse), args.Error(1)
}

func (m *MockProductRepository) GetProductById(ctx context.Context, id string) (*models.ProductResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
}

func (m *MockProductRepository) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductResponse), args.Error(1)
}

func (m *MockProductRepository) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedProductsResponse), args.Error(1)
}

func (m *MockProductRepository) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	args := m.Called(ctx, page, pageSize, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedProductsResponse), args.Error(1)
}

func (m *MockProductRepository) GetProductsCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) GetProductsCountByCategory(ctx context.Context, category string) (int64, error) {
	args := m.Called(ctx, category)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) GetCategories(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProductRepository) CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
}

func (m *MockProductRepository) DeleteProductsByName(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteProductById(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	args := m.Called(ctx, productID, quantityChange)
	return args.Error(0)
}

func (m *MockProductRepository) GetProductQuantity(ctx context.Context, productID string) (int, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockProductRepository) WarmupCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestGetAllProducts(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn []models.ProductResponse
		mockError  error
		wantError  bool
	}{
		{
			name: "success",
			mockReturn: []models.ProductResponse{
				{ID: "1", Name: "Product 1", Price: 10.0},
				{ID: "2", Name: "Product 2", Price: 20.0},
			},
			mockError: nil,
			wantError: false,
		},
		{
			name:       "repository error",
			mockReturn: nil,
			mockError:  errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("GetAllProducts", mock.Anything).Return(tt.mockReturn, tt.mockError)

			service := NewProductService(mockRepo)
			result, err := service.GetAllProducts(context.Background())

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetProductsByName(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		mockReturn  []models.ProductResponse
		mockError   error
		wantError   bool
	}{
		{
			name:        "success",
			productName: "TestProduct",
			mockReturn: []models.ProductResponse{
				{ID: "1", Name: "TestProduct", Price: 10.0},
			},
			mockError: nil,
			wantError: false,
		},
		{
			name:        "repository error",
			productName: "TestProduct",
			mockReturn:  nil,
			mockError:   errors.New("database error"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("GetProductsByName", mock.Anything, tt.productName).Return(tt.mockReturn, tt.mockError)

			service := NewProductService(mockRepo)
			result, err := service.GetProductsByName(context.Background(), tt.productName)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetProductsByPage(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		mockReturn *models.PaginatedProductsResponse
		mockError  error
		wantError  bool
	}{
		{
			name:     "success",
			page:     1,
			pageSize: 10,
			mockReturn: &models.PaginatedProductsResponse{
				Products:   []models.ProductResponse{{ID: "1", Name: "Product 1"}},
				TotalCount: 1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
			mockError: nil,
			wantError: false,
		},
		{
			name:       "repository error",
			page:       1,
			pageSize:   10,
			mockReturn: nil,
			mockError:  errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("GetProductsByPage", mock.Anything, tt.page, tt.pageSize).Return(tt.mockReturn, tt.mockError)

			service := NewProductService(mockRepo)
			result, err := service.GetProductsByPage(context.Background(), tt.page, tt.pageSize)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetProductsByPageAndCategory(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		category   string
		mockReturn *models.PaginatedProductsResponse
		mockError  error
		wantError  bool
	}{
		{
			name:     "success",
			page:     1,
			pageSize: 10,
			category: "Electronics",
			mockReturn: &models.PaginatedProductsResponse{
				Products:   []models.ProductResponse{{ID: "1", Name: "Product 1", Category: "Electronics"}},
				TotalCount: 1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
			mockError: nil,
			wantError: false,
		},
		{
			name:       "repository error",
			page:       1,
			pageSize:   10,
			category:   "Electronics",
			mockReturn: nil,
			mockError:  errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("GetProductsByPageAndCategory", mock.Anything, tt.page, tt.pageSize, tt.category).Return(tt.mockReturn, tt.mockError)

			service := NewProductService(mockRepo)
			result, err := service.GetProductsByPageAndCategory(context.Background(), tt.page, tt.pageSize, tt.category)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetProductsCount(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn int64
		mockError  error
		wantError  bool
	}{
		{
			name:       "success",
			mockReturn: 100,
			mockError:  nil,
			wantError:  false,
		},
		{
			name:       "repository error",
			mockReturn: 0,
			mockError:  errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("GetProductsCount", mock.Anything).Return(tt.mockReturn, tt.mockError)

			service := NewProductService(mockRepo)
			result, err := service.GetProductsCount(context.Background())

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result.Count)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetCategories(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn []string
		mockError  error
		wantError  bool
	}{
		{
			name:       "success",
			mockReturn: []string{"Electronics", "Books", "Clothing"},
			mockError:  nil,
			wantError:  false,
		},
		{
			name:       "repository error",
			mockReturn: nil,
			mockError:  errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("GetCategories", mock.Anything).Return(tt.mockReturn, tt.mockError)

			service := NewProductService(mockRepo)
			result, err := service.GetCategories(context.Background())

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name       string
		request    models.CreateProductRequest
		mockReturn *models.ProductResponse
		mockError  error
		wantError  bool
	}{
		{
			name: "success",
			request: models.CreateProductRequest{
				Name:  "New Product",
				Price: 99.99,
			},
			mockReturn: &models.ProductResponse{
				ID:    "123",
				Name:  "New Product",
				Price: 99.99,
			},
			mockError: nil,
			wantError: false,
		},
		{
			name: "repository error",
			request: models.CreateProductRequest{
				Name:  "New Product",
				Price: 99.99,
			},
			mockReturn: nil,
			mockError:  errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("CreateProduct", mock.Anything, tt.request).Return(tt.mockReturn, tt.mockError)
			if tt.mockError == nil {
				mockRepo.On("GetProductsCount", mock.Anything).Return(int64(1), nil)
			}

			service := NewProductService(mockRepo)
			result, err := service.CreateProduct(context.Background(), tt.request)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteProductsByName(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		mockError   error
		wantError   bool
	}{
		{
			name:        "success",
			productName: "TestProduct",
			mockError:   nil,
			wantError:   false,
		},
		{
			name:        "repository error",
			productName: "TestProduct",
			mockError:   errors.New("database error"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("DeleteProductsByName", mock.Anything, tt.productName).Return(tt.mockError)
			if tt.mockError == nil {
				mockRepo.On("GetProductsCount", mock.Anything).Return(int64(1), nil)
			}

			service := NewProductService(mockRepo)
			err := service.DeleteProductsByName(context.Background(), tt.productName)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteProductById(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		mockError error
		wantError bool
	}{
		{
			name:      "success",
			productID: "123",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "repository error",
			productID: "123",
			mockError: errors.New("database error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			mockRepo.On("DeleteProductById", mock.Anything, tt.productID).Return(tt.mockError)
			if tt.mockError == nil {
				mockRepo.On("GetProductsCount", mock.Anything).Return(int64(1), nil)
			}

			service := NewProductService(mockRepo)
			err := service.DeleteProductById(context.Background(), tt.productID)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateProductQuantity(t *testing.T) {
	tests := []struct {
		name            string
		productID       string
		quantityChange  int
		currentQuantity int
		updateErr       error
		getQuantityErr  error
		wantError       bool
		errorContains   string
	}{
		{
			name:           "success - increase quantity",
			productID:      "123",
			quantityChange: 5,
			updateErr:      nil,
			wantError:      false,
		},
		{
			name:           "success - decrease quantity",
			productID:      "123",
			quantityChange: -5,
			updateErr:      nil,
			wantError:      false,
		},
		{
			name:            "insufficient stock",
			productID:       "123",
			quantityChange:  -15,
			currentQuantity: 10,
			updateErr:       errors.New("insufficient stock or product not found"),
			getQuantityErr:  nil,
			wantError:       true,
			errorContains:   "insufficient stock: current quantity is 10",
		},
		{
			name:           "update error - generic",
			productID:      "123",
			quantityChange: 5,
			updateErr:      errors.New("database error"),
			wantError:      true,
			errorContains:  "database error",
		},
		{
			name:            "insufficient stock - get quantity fails",
			productID:       "123",
			quantityChange:  -15,
			currentQuantity: 0,
			updateErr:       errors.New("insufficient stock or product not found"),
			getQuantityErr:  errors.New("product not found"),
			wantError:       true,
			errorContains:   "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)

			// Expect UpdateProductQuantity to be called first
			mockRepo.On("UpdateProductQuantity", mock.Anything, tt.productID, tt.quantityChange).Return(tt.updateErr)

			// If update fails with specific error, we expect GetProductQuantity
			if tt.updateErr != nil && tt.updateErr.Error() == "insufficient stock or product not found" {
				mockRepo.On("GetProductQuantity", mock.Anything, tt.productID).Return(tt.currentQuantity, tt.getQuantityErr)
			}

			service := NewProductService(mockRepo)
			err := service.UpdateProductQuantity(context.Background(), tt.productID, tt.quantityChange)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncProductCatalogMetric_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockRepo.On("GetProductsCount", mock.Anything).Return(int64(10), nil)

	service := NewProductService(mockRepo)
	service.SyncProductCatalogMetric(context.Background())

	mockRepo.AssertExpectations(t)
}

func TestSyncProductCatalogMetric_Error(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockRepo.On("GetProductsCount", mock.Anything).Return(int64(0), errors.New("count failed"))

	service := NewProductService(mockRepo)
	service.SyncProductCatalogMetric(context.Background())

	mockRepo.AssertExpectations(t)
}

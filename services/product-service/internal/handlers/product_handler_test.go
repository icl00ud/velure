package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"product-service/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductService is a mock implementation of ProductService
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) SyncProductCatalogMetric(ctx context.Context) {}

func (m *MockProductService) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductResponse), args.Error(1)
}

func (m *MockProductService) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductResponse), args.Error(1)
}

func (m *MockProductService) GetProductsByPage(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedProductsResponse), args.Error(1)
}

func (m *MockProductService) GetProductsByPageAndCategory(ctx context.Context, page, pageSize int, category string) (*models.PaginatedProductsResponse, error) {
	args := m.Called(ctx, page, pageSize, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedProductsResponse), args.Error(1)
}

func (m *MockProductService) GetProductsCount(ctx context.Context) (*models.CountResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CountResponse), args.Error(1)
}

func (m *MockProductService) GetCategories(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProductService) CreateProduct(ctx context.Context, product models.CreateProductRequest) (*models.ProductResponse, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
}

func (m *MockProductService) DeleteProductsByName(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockProductService) DeleteProductById(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductService) UpdateProductQuantity(ctx context.Context, productID string, quantityChange int) error {
	args := m.Called(ctx, productID, quantityChange)
	return args.Error(0)
}

func TestSearchProducts(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockReturn     []models.ProductResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:  "success",
			query: "toy",
			mockReturn: []models.ProductResponse{
				{ID: "1", Name: "toy"},
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "missing query",
			query:          "",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "service error",
			query:          "toy",
			mockError:      errors.New("lookup failed"),
			expectedStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			if tt.query != "" {
				mockService.On("GetProductsByName", mock.Anything, tt.query).Return(tt.mockReturn, tt.mockError)
			}

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/products/search", handler.SearchProducts)

			url := "/products/search"
			if tt.query != "" {
				url += "?q=" + tt.query
			}

			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestGetProductsREST(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductsByPage", mock.Anything, 1, 5).Return(&models.PaginatedProductsResponse{
		Products:   []models.ProductResponse{{ID: "1", Name: "item"}},
		Page:       1,
		PageSize:   5,
		TotalCount: 1,
		TotalPages: 1,
	}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products", handler.GetProductsREST)

	req := httptest.NewRequest("GET", "/products?page=1&limit=5", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetAllProducts(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []models.ProductResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			mockReturn: []models.ProductResponse{
				{ID: "1", Name: "Product 1", Price: 10.0},
				{ID: "2", Name: "Product 2", Price: 20.0},
			},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "service error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			mockService.On("GetAllProducts", mock.Anything).Return(tt.mockReturn, tt.mockError)

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/products", handler.GetAllProducts)

			req := httptest.NewRequest("GET", "/products", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetProductsByName(t *testing.T) {
	tests := []struct {
		name           string
		productName    string
		mockReturn     []models.ProductResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "success",
			productName: "TestProduct",
			mockReturn: []models.ProductResponse{
				{ID: "1", Name: "TestProduct", Price: 10.0},
			},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "service error",
			productName:    "TestProduct",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			mockService.On("GetProductsByName", mock.Anything, tt.productName).Return(tt.mockReturn, tt.mockError)

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/products/:name", handler.GetProductsByName)

			req := httptest.NewRequest("GET", "/products/"+tt.productName, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetProductsByPage(t *testing.T) {
	tests := []struct {
		name           string
		page           string
		pageSize       string
		mockReturn     *models.PaginatedProductsResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			page:     "1",
			pageSize: "10",
			mockReturn: &models.PaginatedProductsResponse{
				Products:   []models.ProductResponse{{ID: "1", Name: "Product 1"}},
				TotalCount: 1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "missing page parameter",
			page:           "",
			pageSize:       "10",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name:           "missing pageSize parameter",
			page:           "1",
			pageSize:       "",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name:           "invalid page parameter",
			page:           "invalid",
			pageSize:       "10",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name:           "invalid pageSize parameter",
			page:           "1",
			pageSize:       "invalid",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name:           "service error",
			page:           "1",
			pageSize:       "10",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			if tt.page != "" && tt.pageSize != "" && tt.page != "invalid" && tt.pageSize != "invalid" {
				mockService.On("GetProductsByPage", mock.Anything, 1, 10).Return(tt.mockReturn, tt.mockError)
			}

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/products", handler.GetProductsByPage)

			url := "/products"
			if tt.page != "" || tt.pageSize != "" {
				url += "?page=" + tt.page + "&pageSize=" + tt.pageSize
			}

			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.page != "" && tt.pageSize != "" && tt.page != "invalid" && tt.pageSize != "invalid" {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestGetProductsByPageAndCategory(t *testing.T) {
	tests := []struct {
		name           string
		page           string
		pageSize       string
		category       string
		mockReturn     *models.PaginatedProductsResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			page:     "1",
			pageSize: "10",
			category: "Electronics",
			mockReturn: &models.PaginatedProductsResponse{
				Products:   []models.ProductResponse{{ID: "1", Name: "Product 1", Category: "Electronics"}},
				TotalCount: 1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "missing category",
			page:           "1",
			pageSize:       "10",
			category:       "",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name:           "invalid page",
			page:           "invalid",
			pageSize:       "10",
			category:       "Electronics",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name:           "service error",
			page:           "1",
			pageSize:       "10",
			category:       "Electronics",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			if tt.page == "1" && tt.pageSize == "10" && tt.category != "" {
				mockService.On("GetProductsByPageAndCategory", mock.Anything, 1, 10, tt.category).Return(tt.mockReturn, tt.mockError)
			}

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/products", handler.GetProductsByPageAndCategory)

			url := "/products?page=" + tt.page + "&pageSize=" + tt.pageSize + "&category=" + tt.category
			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.page == "1" && tt.pageSize == "10" && tt.category != "" {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestGetProductsCount(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *models.CountResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			mockReturn:     &models.CountResponse{Count: 100},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "service error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			mockService.On("GetProductsCount", mock.Anything).Return(tt.mockReturn, tt.mockError)

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/count", handler.GetProductsCount)

			req := httptest.NewRequest("GET", "/count", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetCategories(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			mockReturn:     []string{"Electronics", "Books", "Clothing"},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "service error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			mockService.On("GetCategories", mock.Anything).Return(tt.mockReturn, tt.mockError)

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Get("/categories", handler.GetCategories)

			req := httptest.NewRequest("GET", "/categories", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockReturn     *models.ProductResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.CreateProductRequest{
				Name:  "New Product",
				Price: 99.99,
			},
			mockReturn: &models.ProductResponse{
				ID:    "123",
				Name:  "New Product",
				Price: 99.99,
			},
			mockError:      nil,
			expectedStatus: 201,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name: "service error",
			requestBody: models.CreateProductRequest{
				Name:  "New Product",
				Price: 99.99,
			},
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			if req, ok := tt.requestBody.(models.CreateProductRequest); ok {
				mockService.On("CreateProduct", mock.Anything, req).Return(tt.mockReturn, tt.mockError)
			}

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Post("/products", handler.CreateProduct)

			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest("POST", "/products", body)
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestDeleteProductsByName(t *testing.T) {
	tests := []struct {
		name           string
		productName    string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			productName:    "TestProduct",
			mockError:      nil,
			expectedStatus: 204,
		},
		{
			name:           "service error",
			productName:    "TestProduct",
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			mockService.On("DeleteProductsByName", mock.Anything, tt.productName).Return(tt.mockError)

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Delete("/products/:name", handler.DeleteProductsByName)

			req := httptest.NewRequest("DELETE", "/products/"+tt.productName, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteProductById(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			productID:      "123",
			mockError:      nil,
			expectedStatus: 204,
		},
		{
			name:           "service error",
			productID:      "123",
			mockError:      errors.New("database error"),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			mockService.On("DeleteProductById", mock.Anything, tt.productID).Return(tt.mockError)

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Delete("/products/:id", handler.DeleteProductById)

			req := httptest.NewRequest("DELETE", "/products/"+tt.productID, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateProductQuantity(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.UpdateQuantityRequest{
				ProductID:      "123",
				QuantityChange: 5,
			},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name: "missing product ID",
			requestBody: models.UpdateQuantityRequest{
				ProductID:      "",
				QuantityChange: 5,
			},
			mockError:      nil,
			expectedStatus: 400,
		},
		{
			name: "service error",
			requestBody: models.UpdateQuantityRequest{
				ProductID:      "123",
				QuantityChange: 5,
			},
			mockError:      errors.New("insufficient stock"),
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			if req, ok := tt.requestBody.(models.UpdateQuantityRequest); ok && req.ProductID != "" {
				mockService.On("UpdateProductQuantity", mock.Anything, req.ProductID, req.QuantityChange).Return(tt.mockError)
			}

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Post("/quantity", handler.UpdateProductQuantity)

			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest("POST", "/quantity", body)
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

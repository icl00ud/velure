package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"product-service/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockProductService) GetProductById(ctx context.Context, id string) (*models.ProductResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductResponse), args.Error(1)
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

func TestGetProducts_ListAll(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetAllProducts", mock.Anything).Return([]models.ProductResponse{{ID: "1", Name: "p1"}}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products", handler.GetProducts)

	req := httptest.NewRequest("GET", "/products", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetProducts_FilterByName(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductsByName", mock.Anything, "toy").Return([]models.ProductResponse{{ID: "1", Name: "toy"}}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products", handler.GetProducts)

	req := httptest.NewRequest("GET", "/products?name=toy", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetProducts_FilterByQ(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductsByName", mock.Anything, "book").Return([]models.ProductResponse{{ID: "1", Name: "book"}}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products", handler.GetProducts)

	req := httptest.NewRequest("GET", "/products?q=book", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetProducts_Paginated(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductsByPage", mock.Anything, 1, 5).Return(&models.PaginatedProductsResponse{Products: []models.ProductResponse{{ID: "1", Name: "item"}}, Page: 1, PageSize: 5, TotalCount: 1, TotalPages: 1}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products", handler.GetProducts)

	req := httptest.NewRequest("GET", "/products?page=1&limit=5", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetProducts_PaginatedByCategory(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductsByPageAndCategory", mock.Anything, 1, 10, "Electronics").Return(&models.PaginatedProductsResponse{Products: []models.ProductResponse{{ID: "1", Name: "Phone"}}, Page: 1, PageSize: 10, TotalCount: 1, TotalPages: 1}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products", handler.GetProducts)

	req := httptest.NewRequest("GET", "/products?page=1&limit=10&category=Electronics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetProducts_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		url  string
		msg  string
	}{
		{name: "missing limit", url: "/products?page=1", msg: "both page and limit query parameters are required"},
		{name: "invalid page", url: "/products?page=x&limit=10", msg: "Invalid page parameter"},
		{name: "page below minimum", url: "/products?page=0&limit=10", msg: "Invalid page parameter: must be greater than or equal to 1"},
		{name: "invalid limit", url: "/products?page=1&limit=x", msg: "Invalid limit parameter"},
		{name: "limit below minimum", url: "/products?page=1&limit=0", msg: "Invalid limit parameter: must be between 1 and 100"},
		{name: "category without pagination", url: "/products?category=books", msg: "category filter requires page and limit query parameters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewProductHandler(new(MockProductService))
			app := fiber.New()
			app.Get("/products", handler.GetProducts)

			req := httptest.NewRequest("GET", tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, readErr := io.ReadAll(resp.Body)
			assert.NoError(t, readErr)
			assert.Contains(t, string(body), tt.msg)
		})
	}
}

func TestGetProductByID(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductById", mock.Anything, "123").Return(&models.ProductResponse{ID: "123", Name: "item"}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products/:id", handler.GetProductById)

	req := httptest.NewRequest("GET", "/products/123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetProductByID_NotFound(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductById", mock.Anything, "missing").Return((*models.ProductResponse)(nil), errors.New("product not found"))

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products/:id", handler.GetProductById)

	req := httptest.NewRequest("GET", "/products/missing", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetProductsCount(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetProductsCount", mock.Anything).Return(&models.CountResponse{Count: 100}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products/count", handler.GetProductsCount)

	req := httptest.NewRequest("GET", "/products/count", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCategories(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("GetCategories", mock.Anything).Return([]string{"Electronics", "Books"}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Get("/products/categories", handler.GetCategories)

	req := httptest.NewRequest("GET", "/products/categories", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCreateProduct(t *testing.T) {
	mockService := new(MockProductService)
	requestBody := models.CreateProductRequest{Name: "New Product", Price: 99.99}
	mockService.On("CreateProduct", mock.Anything, requestBody).Return(&models.ProductResponse{ID: "123", Name: "New Product", Price: 99.99}, nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Post("/products", handler.CreateProduct)

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteProductByID(t *testing.T) {
	mockService := new(MockProductService)
	mockService.On("DeleteProductById", mock.Anything, "123").Return(nil)

	handler := NewProductHandler(mockService)
	app := fiber.New()
	app.Delete("/products/:id", handler.DeleteProductById)

	req := httptest.NewRequest("DELETE", "/products/123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestPatchProductInventory(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		requestBody    interface{}
		mockError      error
		expectedStatus int
	}{
		{name: "success", pathID: "123", requestBody: map[string]int{"quantity_change": -2}, expectedStatus: fiber.StatusOK},
		{name: "invalid body", pathID: "123", requestBody: "invalid json", expectedStatus: fiber.StatusBadRequest},
		{name: "service error", pathID: "123", requestBody: map[string]int{"quantity_change": -10}, mockError: errors.New("insufficient stock"), expectedStatus: fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockProductService)
			if tt.pathID != "" {
				if payload, ok := tt.requestBody.(map[string]int); ok {
					mockService.On("UpdateProductQuantity", mock.Anything, tt.pathID, payload["quantity_change"]).Return(tt.mockError)
				}
			}

			handler := NewProductHandler(mockService)
			app := fiber.New()
			app.Patch("/products/:id/inventory", handler.PatchProductInventory)

			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest("PATCH", "/products/"+tt.pathID+"/inventory", body)
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestUpdateProduct_NotImplemented(t *testing.T) {
	handler := NewProductHandler(new(MockProductService))
	app := fiber.New()
	app.Put("/products/:id", handler.UpdateProduct)

	req := httptest.NewRequest("PUT", "/products/123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

	body, readErr := io.ReadAll(resp.Body)
	assert.NoError(t, readErr)
	assert.Contains(t, string(body), "product update is not implemented yet")
}

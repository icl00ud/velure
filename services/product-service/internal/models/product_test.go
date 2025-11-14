package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProductModel(t *testing.T) {
	now := time.Now()
	objectID := primitive.NewObjectID()

	product := Product{
		ID:          objectID,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Rating:      4.5,
		Category:    "Electronics",
		Quantity:    10,
		Images:      []string{"image1.jpg", "image2.jpg"},
		Dimensions: Dimensions{
			Height: 10.0,
			Width:  5.0,
			Length: 3.0,
			Weight: 1.5,
		},
		Brand:       "TestBrand",
		Colors:      []string{"Red", "Blue"},
		SKU:         "TEST-123",
		DateCreated: now,
		DateUpdated: now,
	}

	assert.Equal(t, objectID, product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "Test Description", product.Description)
	assert.Equal(t, 99.99, product.Price)
	assert.Equal(t, 4.5, product.Rating)
	assert.Equal(t, "Electronics", product.Category)
	assert.Equal(t, 10, product.Quantity)
	assert.Equal(t, 2, len(product.Images))
	assert.Equal(t, "TestBrand", product.Brand)
	assert.Equal(t, 2, len(product.Colors))
	assert.Equal(t, "TEST-123", product.SKU)
}

func TestDimensions(t *testing.T) {
	dimensions := Dimensions{
		Height: 10.5,
		Width:  5.2,
		Length: 3.8,
		Weight: 1.7,
	}

	assert.Equal(t, 10.5, dimensions.Height)
	assert.Equal(t, 5.2, dimensions.Width)
	assert.Equal(t, 3.8, dimensions.Length)
	assert.Equal(t, 1.7, dimensions.Weight)
}

func TestCreateProductRequest(t *testing.T) {
	request := CreateProductRequest{
		Name:        "New Product",
		Description: "New Description",
		Price:       49.99,
		Rating:      4.0,
		Category:    "Books",
		Quantity:    5,
		Images:      []string{"book.jpg"},
		Dimensions: Dimensions{
			Height: 20.0,
			Width:  15.0,
			Length: 2.0,
			Weight: 0.5,
		},
		Brand:  "Publisher",
		Colors: []string{"White"},
		SKU:    "BOOK-001",
	}

	assert.Equal(t, "New Product", request.Name)
	assert.Equal(t, 49.99, request.Price)
	assert.Equal(t, "Books", request.Category)
	assert.Equal(t, 5, request.Quantity)
	assert.Equal(t, "Publisher", request.Brand)
}

func TestProductResponse(t *testing.T) {
	now := time.Now()

	response := ProductResponse{
		ID:          "507f1f77bcf86cd799439011",
		Name:        "Response Product",
		Description: "Response Description",
		Price:       79.99,
		Rating:      4.8,
		Category:    "Clothing",
		Quantity:    20,
		Images:      []string{"shirt.jpg"},
		Dimensions: Dimensions{
			Height: 30.0,
			Width:  25.0,
			Length: 5.0,
			Weight: 0.3,
		},
		Brand:       "FashionBrand",
		Colors:      []string{"Black", "White"},
		SKU:         "SHIRT-456",
		DateCreated: now,
		DateUpdated: now,
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", response.ID)
	assert.Equal(t, "Response Product", response.Name)
	assert.Equal(t, 79.99, response.Price)
	assert.Equal(t, 4.8, response.Rating)
	assert.Equal(t, "Clothing", response.Category)
	assert.Equal(t, 20, response.Quantity)
}

func TestCountResponse(t *testing.T) {
	count := CountResponse{
		Count: 150,
	}

	assert.Equal(t, int64(150), count.Count)
}

func TestPaginatedProductsResponse(t *testing.T) {
	products := []ProductResponse{
		{ID: "1", Name: "Product 1", Price: 10.0},
		{ID: "2", Name: "Product 2", Price: 20.0},
	}

	response := PaginatedProductsResponse{
		Products:   products,
		TotalCount: 100,
		Page:       1,
		PageSize:   10,
		TotalPages: 10,
	}

	assert.Equal(t, 2, len(response.Products))
	assert.Equal(t, int64(100), response.TotalCount)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 10, response.TotalPages)
}

func TestUpdateQuantityRequest(t *testing.T) {
	request := UpdateQuantityRequest{
		ProductID:      "507f1f77bcf86cd799439011",
		QuantityChange: -5,
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", request.ProductID)
	assert.Equal(t, -5, request.QuantityChange)
}

func TestUpdateQuantityRequest_Positive(t *testing.T) {
	request := UpdateQuantityRequest{
		ProductID:      "123",
		QuantityChange: 10,
	}

	assert.Equal(t, "123", request.ProductID)
	assert.Equal(t, 10, request.QuantityChange)
	assert.True(t, request.QuantityChange > 0)
}

func TestUpdateQuantityRequest_Negative(t *testing.T) {
	request := UpdateQuantityRequest{
		ProductID:      "456",
		QuantityChange: -3,
	}

	assert.Equal(t, "456", request.ProductID)
	assert.Equal(t, -3, request.QuantityChange)
	assert.True(t, request.QuantityChange < 0)
}

func TestDimensions_ZeroValues(t *testing.T) {
	dimensions := Dimensions{}

	assert.Equal(t, 0.0, dimensions.Height)
	assert.Equal(t, 0.0, dimensions.Width)
	assert.Equal(t, 0.0, dimensions.Length)
	assert.Equal(t, 0.0, dimensions.Weight)
}

func TestProduct_WithEmptyFields(t *testing.T) {
	product := Product{
		ID:       primitive.NewObjectID(),
		Name:     "Minimal Product",
		Price:    9.99,
		Quantity: 1,
	}

	assert.NotEmpty(t, product.ID)
	assert.Equal(t, "Minimal Product", product.Name)
	assert.Equal(t, 9.99, product.Price)
	assert.Equal(t, 1, product.Quantity)
	assert.Empty(t, product.Description)
	assert.Empty(t, product.Category)
	assert.Empty(t, product.Brand)
	assert.Empty(t, product.SKU)
}

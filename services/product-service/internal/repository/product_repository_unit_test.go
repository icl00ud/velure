package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"product-service/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestToProductResponse(t *testing.T) {
	repo := &productRepository{}
	now := time.Now()
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        "Dog Toy",
		Description: "Chewy",
		Price:       12.5,
		Rating:      4.2,
		Category:    "Brinquedos",
		Quantity:    5,
		Images:      []string{"img1"},
		Dimensions:  models.Dimensions{Height: 1, Width: 2, Length: 3, Weight: 0.5},
		Brand:       "PetCo",
		Colors:      []string{"Red"},
		SKU:         "SKU123",
		DateCreated: now,
		DateUpdated: now,
	}

	resp := repo.toProductResponse(product)
	if resp.ID != product.ID.Hex() {
		t.Fatalf("expected id %s, got %s", product.ID.Hex(), resp.ID)
	}
	if resp.Name != product.Name || resp.Price != product.Price || resp.Quantity != product.Quantity {
		t.Fatalf("unexpected response mapping: %+v", resp)
	}
	if resp.DateCreated != now || resp.DateUpdated != now {
		t.Fatalf("timestamps not preserved")
	}
}

func TestUpdateProductQuantity_InvalidID_NoDB(t *testing.T) {
	repo := &productRepository{}
	err := repo.UpdateProductQuantity(context.Background(), "bad-id", 1)
	if err == nil {
		t.Fatal("expected error for invalid id")
	}
	if !errors.Is(err, primitive.ErrInvalidHex) && err.Error() == "" {
		t.Fatalf("expected invalid hex error, got %v", err)
	}
}

func TestGetProductQuantity_InvalidID_NoDB(t *testing.T) {
	repo := &productRepository{}
	_, err := repo.GetProductQuantity(context.Background(), "another-bad-id")
	if err == nil {
		t.Fatal("expected error for invalid id")
	}
	if !errors.Is(err, primitive.ErrInvalidHex) && err.Error() == "" {
		t.Fatalf("expected invalid hex error, got %v", err)
	}
}

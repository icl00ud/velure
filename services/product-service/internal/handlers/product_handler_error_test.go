package handlers

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetProductsREST_MissingParams(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{}
	app.Get("/products", handler.GetProductsREST)

	req := httptest.NewRequest("GET", "/products", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetProductsREST_InvalidPage(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{}
	app.Get("/products", handler.GetProductsREST)

	req := httptest.NewRequest("GET", "/products?page=x&limit=10", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetProductsByPage_MissingParams(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{}
	app.Get("/products-legacy", handler.GetProductsByPage)

	req := httptest.NewRequest("GET", "/products-legacy?page=1", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetProductsByPage_Error(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{err: errors.New("db down")}
	app.Get("/products-legacy", handler.GetProductsByPage)

	req := httptest.NewRequest("GET", "/products-legacy?page=1&pageSize=10", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 500 {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestCreateProduct_BodyError(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{}
	app.Post("/products", handler.CreateProduct)

	req := httptest.NewRequest("POST", "/products", strings.NewReader("bad"))
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateProduct_ServiceError(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{err: errors.New("fail")}
	app.Post("/products", handler.CreateProduct)

	body := `{"name":"p","price":10}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	resp, _ := app.Test(req)
	if resp.StatusCode != 500 && resp.StatusCode != 400 {
		t.Fatalf("expected 400/500, got %d", resp.StatusCode)
	}
}

func TestUpdateProductQuantity_MissingID(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{}
	app.Put("/quantity", handler.UpdateProductQuantity)

	body := `{"product_id":"","quantity_change":-1}`
	req := httptest.NewRequest("PUT", "/quantity", strings.NewReader(body))
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateProductQuantity_ServiceError(t *testing.T) {
	app, handler := setupTestApp()
	handler.service = &stubProductService{err: errors.New("fail")}
	app.Put("/quantity", handler.UpdateProductQuantity)

	body := `{"product_id":"507f1f77bcf86cd799439011","quantity_change":-1}`
	req := httptest.NewRequest("PUT", "/quantity", strings.NewReader(body))
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

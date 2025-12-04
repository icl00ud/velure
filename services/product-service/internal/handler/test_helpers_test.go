package handlers

import (
	"context"
	"errors"
	"testing"

	"product-service/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestStubProductService_ReturnsConfiguredError(t *testing.T) {
	errExpected := errors.New("boom")
	svc := &stubProductService{err: errExpected}

	svc.SyncProductCatalogMetric(context.Background())

	_, err := svc.GetAllProducts(context.Background())
	assert.Equal(t, errExpected, err)

	_, err = svc.GetProductsByName(context.Background(), "toy")
	assert.Equal(t, errExpected, err)

	_, err = svc.GetProductsByPage(context.Background(), 1, 10)
	assert.Equal(t, errExpected, err)

	_, err = svc.GetProductsByPageAndCategoryFromCache(context.Background(), 1, 10, "cats")
	assert.Equal(t, errExpected, err)

	_, err = svc.GetProductsCountByCategory(context.Background(), "dogs")
	assert.Equal(t, errExpected, err)

	_, err = svc.CreateProduct(context.Background(), models.CreateProductRequest{Name: "name"})
	assert.Equal(t, errExpected, err)

	err = svc.UpdateProductQuantity(context.Background(), "1", 1)
	assert.Equal(t, errExpected, err)

	_, err = svc.GetProductQuantity(context.Background(), "1")
	assert.Equal(t, errExpected, err)
}

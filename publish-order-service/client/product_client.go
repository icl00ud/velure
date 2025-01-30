package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/icl00ud/publish-order-service/domain"
)

type ProductClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (pc *ProductClient) GetProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/products/%s", pc.BaseURL, productID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var product domain.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode product response: %w", err)
	}

	return &product, nil
}

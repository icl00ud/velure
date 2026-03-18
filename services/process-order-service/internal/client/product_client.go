package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PermanentError struct {
	Message    string
	StatusCode int
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("permanent error (%d): %s", e.StatusCode, e.Message)
}

type TransientError struct {
	Message    string
	StatusCode int
}

func (e *TransientError) Error() string {
	return fmt.Sprintf("transient error (%d): %s", e.StatusCode, e.Message)
}

type ProductClient interface {
	UpdateQuantity(productID string, quantityChange int) error
}

type productClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string) ProductClient {
	return &productClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type UpdateQuantityRequest struct {
	QuantityChange int `json:"quantity_change"`
}

func (c *productClient) UpdateQuantity(productID string, quantityChange int) error {
	req := UpdateQuantityRequest{
		QuantityChange: quantityChange,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("PATCH", c.baseURL+"/api/products/"+productID+"/inventory", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &TransientError{
			Message:    err.Error(),
			StatusCode: 0,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		errMsg := fmt.Sprintf("status %d", resp.StatusCode)
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			errMsg = errResp.Error
		}

		switch {
		case resp.StatusCode == http.StatusBadRequest,
			resp.StatusCode == http.StatusNotFound,
			resp.StatusCode == http.StatusConflict,
			resp.StatusCode == http.StatusUnprocessableEntity:
			return &PermanentError{
				Message:    errMsg,
				StatusCode: resp.StatusCode,
			}

		case resp.StatusCode == http.StatusTooManyRequests,
			resp.StatusCode >= 500:
			return &TransientError{
				Message:    errMsg,
				StatusCode: resp.StatusCode,
			}

		default:
			return &PermanentError{
				Message:    errMsg,
				StatusCode: resp.StatusCode,
			}
		}
	}

	return nil
}

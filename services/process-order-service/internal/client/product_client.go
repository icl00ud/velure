package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PermanentError indica erro permanente que não deve ser retryado (ex: produto não encontrado)
type PermanentError struct {
	Message    string
	StatusCode int
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("permanent error (%d): %s", e.StatusCode, e.Message)
}

// TransientError indica erro temporário que pode ser retryado (ex: timeout, 5xx)
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
	ProductID      string `json:"product_id"`
	QuantityChange int    `json:"quantity_change"`
}

func (c *productClient) UpdateQuantity(productID string, quantityChange int) error {
	req := UpdateQuantityRequest{
		ProductID:      productID,
		QuantityChange: quantityChange,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/product/updateQuantity", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Erros de rede/timeout são temporários - podem ser retryados
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

		// Classificar erro baseado no status code HTTP
		switch {
		// Erros permanentes 4xx (client errors) - NÃO devem ser retryados
		case resp.StatusCode == http.StatusBadRequest,
			resp.StatusCode == http.StatusNotFound,
			resp.StatusCode == http.StatusConflict,
			resp.StatusCode == http.StatusUnprocessableEntity:
			return &PermanentError{
				Message:    errMsg,
				StatusCode: resp.StatusCode,
			}

		// Erros temporários 5xx ou 429 - PODEM ser retryados
		case resp.StatusCode == http.StatusTooManyRequests,
			resp.StatusCode >= 500:
			return &TransientError{
				Message:    errMsg,
				StatusCode: resp.StatusCode,
			}

		// Outros erros não esperados - tratar como permanentes por segurança
		default:
			return &PermanentError{
				Message:    errMsg,
				StatusCode: resp.StatusCode,
			}
		}
	}

	return nil
}

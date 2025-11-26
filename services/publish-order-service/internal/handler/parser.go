package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/icl00ud/publish-order-service/internal/model"
)

type createOrderDTO struct {
	Items []model.CartItem `json:"items"`
}

func parseCreateOrder(r io.Reader) ([]model.CartItem, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Try parsing as { "items": [...] } first
	var dto createOrderDTO
	if err := json.Unmarshal(body, &dto); err == nil && len(dto.Items) > 0 {
		return dto.Items, nil
	}

	// Fallback: try parsing as [...]
	var items []model.CartItem
	if err := json.Unmarshal(body, &items); err == nil {
		return items, nil
	}

	// If both fail, return the error from the first attempt (or a generic one)
	return nil, json.Unmarshal(body, &dto)
}

func parsePagination(r *http.Request) (page, pageSize int) {
	page = 1
	pageSize = 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	return page, pageSize
}

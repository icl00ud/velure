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
	var dto createOrderDTO
	if err := json.NewDecoder(r).Decode(&dto); err != nil {
		return nil, err
	}
	return dto.Items, nil
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

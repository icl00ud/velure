package handler

import (
	"encoding/json"
	"io"

	"github.com/icl00ud/publish-order-service/pkg/model"
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

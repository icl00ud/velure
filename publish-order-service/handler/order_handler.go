package handler

import (
	"fmt"

	"github.com/icl00ud/velure-order-service/domain"
	"github.com/icl00ud/velure-order-service/queue"
	"gofr.dev/pkg/gofr"
)

type OrderHandler struct {
	rabbitRepo *queue.RabbitMQRepository
}

func NewOrderHandler(rabbitRepo *queue.RabbitMQRepository) *OrderHandler {
	return &OrderHandler{
		rabbitRepo: rabbitRepo,
	}
}

func (h *OrderHandler) CreateOrder(ctx *gofr.Context) (any, error) {
	var order domain.Order
	if err := ctx.Bind(&order); err != nil {
		return nil, err
	}

	ctx.Logger.Info(fmt.Sprintf("Order: %v", order))
	event := domain.Event{
		Type:  domain.OrderCreated,
		Order: order,
	}

	if err := h.rabbitRepo.PublishEvent(event); err != nil {
		return nil, err
	}

	return order, nil
}

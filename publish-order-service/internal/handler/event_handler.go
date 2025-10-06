package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
)

type EventHandler struct {
	orderService *service.OrderService
	logger       *zap.Logger
}

func NewEventHandler(orderService *service.OrderService, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		orderService: orderService,
		logger:       logger,
	}
}

func (h *EventHandler) HandleEvent(ctx context.Context, evt model.Event) error {
	switch evt.Type {
	case model.OrderProcessing:
		return h.handleOrderProcessing(ctx, evt.Payload)
	case model.OrderCompleted:
		return h.handleOrderCompleted(ctx, evt.Payload)
	default:
		h.logger.Warn("unhandled event type", zap.String("type", evt.Type))
		return nil
	}
}

func (h *EventHandler) handleOrderProcessing(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		h.logger.Error("failed to unmarshal order.processing payload", zap.Error(err))
		return fmt.Errorf("unmarshal processing payload: %w", err)
	}

	if data.ID == "" {
		return fmt.Errorf("order id is empty")
	}

	h.logger.Info("updating order status to PROCESSING", zap.String("order_id", data.ID))

	if _, err := h.orderService.UpdateStatus(ctx, data.ID, model.StatusProcessing); err != nil {
		h.logger.Error("failed to update order status to PROCESSING",
			zap.String("order_id", data.ID),
			zap.Error(err))
		return fmt.Errorf("update status to processing: %w", err)
	}

	h.logger.Info("order status updated to PROCESSING", zap.String("order_id", data.ID))
	return nil
}

func (h *EventHandler) handleOrderCompleted(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		ID      string `json:"id"`
		OrderID string `json:"order_id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		h.logger.Error("failed to unmarshal order.completed payload", zap.Error(err))
		return fmt.Errorf("unmarshal completed payload: %w", err)
	}

	orderID := data.OrderID
	if orderID == "" {
		orderID = data.ID
	}

	if orderID == "" {
		return fmt.Errorf("order id is empty")
	}

	h.logger.Info("updating order status to COMPLETED", zap.String("order_id", orderID))

	if _, err := h.orderService.UpdateStatus(ctx, orderID, model.StatusCompleted); err != nil {
		h.logger.Error("failed to update order status to COMPLETED",
			zap.String("order_id", orderID),
			zap.Error(err))
		return fmt.Errorf("update status to completed: %w", err)
	}

	h.logger.Info("order status updated to COMPLETED", zap.String("order_id", orderID))
	return nil
}

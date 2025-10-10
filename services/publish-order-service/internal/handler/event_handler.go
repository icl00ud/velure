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
	sseHandler   *SSEHandler
}

func NewEventHandler(orderService *service.OrderService, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		orderService: orderService,
		logger:       logger,
	}
}

func (h *EventHandler) SetSSEHandler(sseHandler *SSEHandler) {
	h.sseHandler = sseHandler
}

func (h *EventHandler) HandleEvent(ctx context.Context, evt model.Event) error {
	zap.L().Info("event received", zap.String("type", evt.Type))

	if evt.Type == model.OrderProcessing || evt.Type == model.OrderCompleted {
		var payload struct {
			ID      string `json:"id"`
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal event payload: %w", err)
		}

		// Use OrderID if available, otherwise use ID
		orderID := payload.OrderID
		if orderID == "" {
			orderID = payload.ID
		}

		if orderID == "" {
			return fmt.Errorf("order id is empty in event payload")
		}

		// Map event type to status
		status := model.StatusProcessing
		if evt.Type == model.OrderCompleted {
			status = model.StatusCompleted
		}

		order, err := h.orderService.UpdateStatus(ctx, orderID, status)
		if err != nil {
			return fmt.Errorf("update status: %w", err)
		}
		zap.L().Info("order status updated", zap.String("order_id", orderID), zap.String("status", status))

		if h.sseHandler != nil {
			h.sseHandler.NotifyOrderUpdate(order)
		}
	}

	return nil
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

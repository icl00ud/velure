package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/icl00ud/velure-shared/logger"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
)

type EventHandler struct {
	orderService *service.OrderService
	logger       *logger.Logger
	sseHandler   *SSEHandler
}

func NewEventHandler(orderService *service.OrderService, log *logger.Logger) *EventHandler {
	return &EventHandler{
		orderService: orderService,
		logger:       log,
	}
}

func (h *EventHandler) SetSSEHandler(sseHandler *SSEHandler) {
	h.sseHandler = sseHandler
}

func (h *EventHandler) HandleEvent(ctx context.Context, evt model.Event) error {
	logger.Info("event received", logger.String("type", evt.Type))

	if evt.Type == model.OrderProcessing || evt.Type == model.OrderCompleted || evt.Type == model.OrderFailed {
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
		var status string
		switch evt.Type {
		case model.OrderProcessing:
			status = model.StatusProcessing
		case model.OrderCompleted:
			status = model.StatusCompleted
		case model.OrderFailed:
			status = model.StatusFailed
		}

		order, err := h.orderService.UpdateStatus(ctx, orderID, status)
		if err != nil {
			// If order not found, it was likely deleted - just ACK the message
			if err.Error() == "sql: no rows in result set" {
				logger.Warn("order not found in database, skipping status update",
					logger.String("order_id", orderID),
					logger.String("status", status))
				return nil
			}
			return fmt.Errorf("update status: %w", err)
		}
		logger.Info("order status updated", logger.String("order_id", orderID), logger.String("status", status))

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
		h.logger.Error("failed to unmarshal order.processing payload", logger.Err(err))
		return fmt.Errorf("unmarshal processing payload: %w", err)
	}

	if data.ID == "" {
		return fmt.Errorf("order id is empty")
	}

	h.logger.Info("updating order status to PROCESSING", logger.String("order_id", data.ID))

	if _, err := h.orderService.UpdateStatus(ctx, data.ID, model.StatusProcessing); err != nil {
		h.logger.Error("failed to update order status to PROCESSING",
			logger.String("order_id", data.ID),
			logger.Err(err))
		return fmt.Errorf("update status to processing: %w", err)
	}

	h.logger.Info("order status updated to PROCESSING", logger.String("order_id", data.ID))
	return nil
}

func (h *EventHandler) handleOrderCompleted(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		ID      string `json:"id"`
		OrderID string `json:"order_id"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		h.logger.Error("failed to unmarshal order.completed payload", logger.Err(err))
		return fmt.Errorf("unmarshal completed payload: %w", err)
	}

	orderID := data.OrderID
	if orderID == "" {
		orderID = data.ID
	}

	if orderID == "" {
		return fmt.Errorf("order id is empty")
	}

	h.logger.Info("updating order status to COMPLETED", logger.String("order_id", orderID))

	if _, err := h.orderService.UpdateStatus(ctx, orderID, model.StatusCompleted); err != nil {
		h.logger.Error("failed to update order status to COMPLETED",
			logger.String("order_id", orderID),
			logger.Err(err))
		return fmt.Errorf("update status to completed: %w", err)
	}

	h.logger.Info("order status updated to COMPLETED", logger.String("order_id", orderID))
	return nil
}

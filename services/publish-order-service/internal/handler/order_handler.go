package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/icl00ud/publish-order-service/internal/metrics"
	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
	"github.com/icl00ud/velure-shared/logger"
)

type Publisher interface {
	Publish(evt model.Event) error
}

type OrderHandler struct {
	svc OrderService
	pub Publisher
}

func NewOrderHandler(svc OrderService, pub Publisher) *OrderHandler {
	return &OrderHandler{svc: svc, pub: pub}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		logger.Warn("missing user_id in context")
		metrics.HTTPRequests.WithLabelValues("publish-order-service", "POST", "/orders", "401").Inc()
		metrics.HTTPRequestDuration.WithLabelValues("publish-order-service", "POST", "/orders").Observe(time.Since(start).Seconds())
		writeJSON(w, http.StatusUnauthorized, response{"error": "unauthorized"})
		return
	}

	items, err := parseCreateOrder(r.Body)
	if err != nil {
		logger.Warn("invalid payload", logger.Err(err))
		metrics.HTTPRequests.WithLabelValues("publish-order-service", "POST", "/orders", "400").Inc()
		metrics.HTTPRequestDuration.WithLabelValues("publish-order-service", "POST", "/orders").Observe(time.Since(start).Seconds())
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid payload"})
		return
	}

	o, err := h.svc.Create(r.Context(), userID, items)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, service.ErrNoItems) || errors.Is(err, service.ErrInvalidItem) {
			code = http.StatusBadRequest
		}
		logger.Error("create order failed", logger.Err(err))
		metrics.OrdersCreated.WithLabelValues("failure").Inc()
		metrics.HTTPRequests.WithLabelValues("publish-order-service", "POST", "/orders", http.StatusText(code)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues("publish-order-service", "POST", "/orders").Observe(time.Since(start).Seconds())
		writeJSON(w, code, response{"error": err.Error()})
		return
	}

	metrics.OrdersCreated.WithLabelValues("success").Inc()
	metrics.OrderCreationDuration.Observe(time.Since(start).Seconds())
	metrics.OrderTotalValue.Observe(float64(o.Total))
	metrics.OrderItemsCount.Observe(float64(len(o.Items)))

	evt := model.Event{Type: model.OrderCreated, Payload: mustMarshal(o)}
	if err := h.pub.Publish(evt); err != nil {
		logger.Error("publish event failed", logger.Err(err))
		metrics.Errors.WithLabelValues("rabbitmq").Inc()
		metrics.OrdersPublished.WithLabelValues("failure").Inc()
	} else {
		metrics.OrdersPublished.WithLabelValues("success").Inc()
	}

	metrics.HTTPRequests.WithLabelValues("publish-order-service", "POST", "/orders", "201").Inc()
	metrics.HTTPRequestDuration.WithLabelValues("publish-order-service", "POST", "/orders").Observe(time.Since(start).Seconds())
	writeJSON(w, http.StatusCreated, response{
		"order_id": o.ID,
		"total":    o.Total,
		"status":   o.Status,
	})
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		logger.Warn("missing user_id in context")
		writeJSON(w, http.StatusUnauthorized, response{"error": "unauthorized"})
		return
	}

	page, pageSize := parsePagination(r)

	result, err := h.svc.GetOrdersByUserID(r.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("get user orders failed", logger.Err(err))
		writeJSON(w, http.StatusInternalServerError, response{"error": "internal error"})
		return
	}

	writeJSONData(w, http.StatusOK, result)
}

func (h *OrderHandler) GetUserOrderByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		logger.Warn("missing user_id in context")
		writeJSON(w, http.StatusUnauthorized, response{"error": "unauthorized"})
		return
	}

	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		logger.Warn("missing order_id in query")
		writeJSON(w, http.StatusBadRequest, response{"error": "order_id required"})
		return
	}

	order, err := h.svc.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		logger.Error("get user order by id failed", logger.Err(err))
		writeJSON(w, http.StatusNotFound, response{"error": "order not found"})
		return
	}

	writeJSONData(w, http.StatusOK, order)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var dto struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		logger.Warn("invalid payload", logger.Err(err))
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid payload"})
		return
	}

	o, err := h.svc.UpdateStatus(r.Context(), dto.OrderID, dto.Status)
	if err != nil {
		logger.Error("update status failed", logger.Err(err))
		writeJSON(w, http.StatusInternalServerError, response{"error": "internal error"})
		return
	}

	evt := model.Event{Type: dto.Status, Payload: mustMarshal(o)}
	if err := h.pub.Publish(evt); err != nil {
		logger.Error("publish event failed", logger.Err(err))
	}

	writeJSON(w, http.StatusOK, response{
		"order_id": o.ID,
		"status":   o.Status,
		"updated":  o.UpdatedAt,
	})
}

func (h *OrderHandler) GetOrdersByPage(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)

	result, err := h.svc.GetOrdersByPage(r.Context(), page, pageSize)
	if err != nil {
		logger.Error("get orders by page failed", logger.Err(err))
		writeJSON(w, http.StatusInternalServerError, response{"error": "internal error"})
		return
	}

	writeJSONData(w, http.StatusOK, result)
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

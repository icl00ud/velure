package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/publish-order-service/internal/middleware"
	"github.com/icl00ud/velure/services/publish-order-service/internal/service"
	"github.com/icl00ud/velure/shared/logger"
)

type OrderHandler struct {
	svc OrderService
}

func NewOrderHandler(svc OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// trackCreateOrder records the request metrics for POST /orders once per
// request, regardless of which branch responded.
func trackCreateOrder(status string, start time.Time) {
	metrics.HTTPRequests.WithLabelValues("publish-order-service", "POST", "/orders", status).Inc()
	metrics.HTTPRequestDuration.WithLabelValues("publish-order-service", "POST", "/orders").Observe(time.Since(start).Seconds())
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		logger.Warn("missing user_id in context")
		trackCreateOrder("401", start)
		writeJSON(w, http.StatusUnauthorized, response{"error": "unauthorized"})
		return
	}

	items, err := parseCreateOrder(r.Body)
	if err != nil {
		logger.Warn("invalid payload", logger.Err(err))
		trackCreateOrder("400", start)
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid payload"})
		return
	}

	o, err := h.svc.Create(r.Context(), userID, items)
	if err != nil {
		code := http.StatusInternalServerError
		// Validation errors are safe to echo back; anything else stays generic
		// so internals (SQL, broker state) never reach the client.
		msg := "internal error"
		if errors.Is(err, service.ErrNoItems) || errors.Is(err, service.ErrInvalidItem) {
			code = http.StatusBadRequest
			msg = err.Error()
		}
		logger.Error("create order failed", logger.Err(err))
		metrics.OrdersCreated.WithLabelValues("failure").Inc()
		trackCreateOrder(http.StatusText(code), start)
		writeJSON(w, code, response{"error": msg})
		return
	}

	metrics.OrdersCreated.WithLabelValues("success").Inc()
	metrics.OrderCreationDuration.Observe(time.Since(start).Seconds())
	metrics.OrderTotalValue.Observe(float64(o.Total))
	metrics.OrderItemsCount.Observe(float64(len(o.Items)))

	trackCreateOrder("201", start)
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



package handler

import (
	"encoding/json"
	"net/http"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
	"go.uber.org/zap"
)

type Publisher interface {
	Publish(evt model.Event) error
}

type OrderHandler struct {
	svc *service.OrderService
	pub Publisher
}

func NewOrderHandler(svc *service.OrderService, pub Publisher) *OrderHandler {
	return &OrderHandler{svc: svc, pub: pub}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	items, err := parseCreateOrder(r.Body)
	if err != nil {
		zap.L().Warn("invalid payload", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid payload"})
		return
	}

	o, err := h.svc.Create(r.Context(), items)
	if err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNoItems {
			code = http.StatusBadRequest
		}
		zap.L().Error("create order failed", zap.Error(err))
		writeJSON(w, code, response{"error": err.Error()})
		return
	}

	evt := model.Event{Type: model.OrderCreated, Payload: mustMarshal(o)}
	if err := h.pub.Publish(evt); err != nil {
		zap.L().Error("publish event failed", zap.Error(err))
	}

	writeJSON(w, http.StatusCreated, response{
		"order_id": o.ID,
		"total":    o.Total,
		"status":   o.Status,
	})
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var dto struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		zap.L().Warn("invalid payload", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid payload"})
		return
	}

	o, err := h.svc.UpdateStatus(r.Context(), dto.OrderID, dto.Status)
	if err != nil {
		zap.L().Error("update status failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, response{"error": "internal error"})
		return
	}

	evt := model.Event{Type: dto.Status, Payload: mustMarshal(o)}
	if err := h.pub.Publish(evt); err != nil {
		zap.L().Error("publish event failed", zap.Error(err))
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
		zap.L().Error("get orders by page failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, response{"error": "internal error"})
		return
	}

	writeJSONData(w, http.StatusOK, result)
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

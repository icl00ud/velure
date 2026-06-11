package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/publish-order-service/internal/middleware"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
)

type SSEHandler struct {
	svc      OrderService
	registry *SSERegistry
	bus      OrderUpdateBus
}

func NewSSEHandler(svc OrderService) *SSEHandler {
	return &SSEHandler{
		svc:      svc,
		registry: NewSSERegistry(),
	}
}

// AttachBus wires a cross-replica update bus. Must be called before StartBus
// and before the handler starts serving.
func (h *SSEHandler) AttachBus(bus OrderUpdateBus) {
	h.bus = bus
}

// StartBus subscribes to the bus and forwards remote updates to the local
// registry. No-op when no bus is attached.
func (h *SSEHandler) StartBus(ctx context.Context) error {
	if h.bus == nil {
		return nil
	}
	return h.bus.Subscribe(ctx, func(order model.Order) {
		h.registry.Broadcast(order.ID, order)
	})
}

func (h *SSEHandler) StreamOrderStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		logger.Warn("missing user_id in context")
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		logger.Warn("missing order_id in query")
		http.Error(w, `{"error":"order_id required"}`, http.StatusBadRequest)
		return
	}

	order, err := h.svc.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		logger.Error("get user order by id failed", logger.Err(err))
		http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error("streaming not supported")
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events := make(chan model.Order, 10)
	h.registry.Register(orderID, events)
	defer h.registry.Unregister(orderID, events)

	metrics.SSEConnections.Inc()
	defer metrics.SSEConnections.Dec()

	sendEvent := func(o model.Order) error {
		data, err := json.Marshal(o)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			return err
		}
		flusher.Flush()
		metrics.SSEMessagesSent.Inc()
		return nil
	}

	if err := sendEvent(order); err != nil {
		logger.Error("failed to send initial event", logger.Err(err))
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			logger.Info("client disconnected", logger.String("order_id", orderID))
			return
		case updatedOrder := <-events:
			if err := sendEvent(updatedOrder); err != nil {
				logger.Error("failed to send event", logger.Err(err))
				return
			}
		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": keepalive\n\n")
			if err != nil {
				logger.Error("failed to send keepalive", logger.Err(err))
				return
			}
			flusher.Flush()
		}
	}
}

// NotifyOrderUpdate distributes an order update to SSE subscribers. With a
// bus attached the update goes through Redis (and comes back to the local
// registry via the subscription, together with every other replica); without
// one it is broadcast locally.
func (h *SSEHandler) NotifyOrderUpdate(order model.Order) {
	if h.bus != nil {
		if err := h.bus.Publish(context.Background(), order); err != nil {
			logger.Error("bus publish failed, falling back to local broadcast", logger.Err(err))
			h.registry.Broadcast(order.ID, order)
		}
		return
	}
	h.registry.Broadcast(order.ID, order)
}

type SSERegistry struct {
	mu          sync.RWMutex
	subscribers map[string][]chan model.Order
}

func NewSSERegistry() *SSERegistry {
	return &SSERegistry{
		subscribers: make(map[string][]chan model.Order),
	}
}

func (r *SSERegistry) Register(orderID string, ch chan model.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.subscribers[orderID] == nil {
		r.subscribers[orderID] = []chan model.Order{}
	}
	r.subscribers[orderID] = append(r.subscribers[orderID], ch)
}

func (r *SSERegistry) Unregister(orderID string, ch chan model.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subs := r.subscribers[orderID]
	for i, sub := range subs {
		if sub == ch {
			r.subscribers[orderID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(r.subscribers[orderID]) == 0 {
		delete(r.subscribers, orderID)
	}
}

func (r *SSERegistry) Broadcast(orderID string, order model.Order) {
	r.mu.RLock()
	subs := make([]chan model.Order, len(r.subscribers[orderID]))
	copy(subs, r.subscribers[orderID])
	r.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- order:
		default:
			logger.Warn("channel full, dropping event", logger.String("order_id", orderID))
		}
	}
}

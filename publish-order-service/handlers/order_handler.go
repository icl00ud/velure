package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/icl00ud/publish-order-service/domain"
	"github.com/icl00ud/publish-order-service/queue"
	"github.com/icl00ud/publish-order-service/storage"
)

type OrderHandler struct {
	Storage    *storage.Storage
	RabbitRepo *queue.RabbitMQRepository
}

func NewOrderHandler(s *storage.Storage, rr *queue.RabbitMQRepository) *OrderHandler {
	return &OrderHandler{
		Storage:    s,
		RabbitRepo: rr,
	}
}

func (oh *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderInput struct {
		Items []struct {
			ProductID string  `json:"product_id"`
			Name      string  `json:"name"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orderInput); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(orderInput.Items) == 0 {
		http.Error(w, "No items in the cart", http.StatusBadRequest)
		return
	}

	var cartItems []domain.CartItem
	total := 0

	// Uso as informações recebidas para montar os itens do pedido
	for _, item := range orderInput.Items {
		cartItems = append(cartItems, domain.CartItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
		total += int(item.Price * float64(item.Quantity))
	}

	orderID := uuid.New().String()

	order := domain.Order{
		ID:        orderID,
		Items:     cartItems,
		Total:     total,
		Status:    string(domain.OrderCreated),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	oh.Storage.CreateOrder(order)

	eventPayload, err := json.Marshal(order)
	if err != nil {
		log.Printf("Error marshaling event payload: %v", err)
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	event := domain.Event{
		Type:    domain.OrderCreated,
		Payload: eventPayload,
	}

	oh.RabbitRepo.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": order.ID,
		"total":    order.Total,
		"status":   order.Status,
	})
}

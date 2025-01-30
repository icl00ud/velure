// handlers/order_handler.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/icl00ud/publish-order-service/client"
	"github.com/icl00ud/publish-order-service/domain"
	"github.com/icl00ud/publish-order-service/queue"
	"github.com/icl00ud/publish-order-service/storage"
)

type OrderHandler struct {
	ProductClient *client.ProductClient
	Storage       *storage.Storage
	RabbitRepo    *queue.RabbitMQRepository
}

func NewOrderHandler(pc *client.ProductClient, s *storage.Storage, rr *queue.RabbitMQRepository) *OrderHandler {
	return &OrderHandler{
		ProductClient: pc,
		Storage:       s,
		RabbitRepo:    rr,
	}
}

func (oh *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderInput struct {
		Items []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
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

	// Buscar detalhes dos produtos no Product Service
	for _, item := range orderInput.Items {
		product, err := oh.ProductClient.GetProductByID(r.Context(), item.ProductID)
		if err != nil {
			log.Printf("Error fetching product %s: %v", item.ProductID, err)
			http.Error(w, "Failed to fetch product details", http.StatusInternalServerError)
			return
		}

		if !product.Disponibility {
			http.Error(w, "Product "+product.Name+" is not available", http.StatusBadRequest)
			return
		}

		if product.QuantityWarehouse < item.Quantity {
			http.Error(w, "Insufficient quantity for product "+product.Name, http.StatusBadRequest)
			return
		}

		cartItems = append(cartItems, domain.CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})

		total += item.Quantity * product.Price
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

	if err := oh.Storage.CreateOrder(order); err != nil {
		log.Printf("Error saving order to database: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Publica o evento OrderCreated no RabbitMQ
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

	if err := oh.RabbitRepo.PublishEvent(event); err != nil {
		log.Printf("Error publishing event to RabbitMQ: %v", err)
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	// Retorna a resposta com os detalhes do pedido
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": order.ID,
		"total":    order.Total,
		"status":   order.Status,
	})
}

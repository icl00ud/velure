package handler

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
)

// Two SSEHandlers backed by the same Redis simulate two replicas of the
// service: an update notified on replica A must reach a subscriber whose SSE
// connection lives on replica B.
func TestSSEBus_UpdateCrossesReplicas(t *testing.T) {
	mr := miniredis.RunT(t)

	clientA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer clientA.Close()
	clientB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer clientB.Close()

	replicaA := NewSSEHandler(&routeStubService{})
	replicaA.AttachBus(NewRedisOrderBus(clientA))
	replicaB := NewSSEHandler(&routeStubService{})
	replicaB.AttachBus(NewRedisOrderBus(clientB))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := replicaA.StartBus(ctx); err != nil {
		t.Fatalf("start bus A: %v", err)
	}
	if err := replicaB.StartBus(ctx); err != nil {
		t.Fatalf("start bus B: %v", err)
	}

	// Client connected to replica B.
	events := make(chan model.Order, 10)
	replicaB.registry.Register("order-x", events)
	defer replicaB.registry.Unregister("order-x", events)

	// Status update consumed by replica A.
	replicaA.NotifyOrderUpdate(model.Order{ID: "order-x", Status: "COMPLETED"})

	select {
	case got := <-events:
		if got.ID != "order-x" || got.Status != "COMPLETED" {
			t.Fatalf("unexpected order delivered: %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("update published on replica A never reached subscriber on replica B")
	}
}

// Without a bus the handler must keep working as a single replica.
func TestSSEHandler_NoBusFallsBackToLocalBroadcast(t *testing.T) {
	h := NewSSEHandler(&routeStubService{})

	events := make(chan model.Order, 10)
	h.registry.Register("order-y", events)
	defer h.registry.Unregister("order-y", events)

	h.NotifyOrderUpdate(model.Order{ID: "order-y", Status: "PROCESSING"})

	select {
	case got := <-events:
		if got.Status != "PROCESSING" {
			t.Fatalf("unexpected order: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("local broadcast did not deliver")
	}
}

type routeStubService struct{}

func (s *routeStubService) Create(_ context.Context, userID string, items []model.CartItem) (model.Order, error) {
	return model.Order{}, nil
}
func (s *routeStubService) UpdateStatus(_ context.Context, id, status string) (model.Order, error) {
	return model.Order{ID: id, Status: status}, nil
}
func (s *routeStubService) GetOrdersByUserID(_ context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return nil, nil
}
func (s *routeStubService) GetOrderByID(_ context.Context, userID, orderID string) (model.Order, error) {
	return model.Order{ID: orderID}, nil
}

package handler

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
)

// orderUpdatesChannel is the Redis pub/sub channel that fans order status
// updates out to every service replica.
const orderUpdatesChannel = "order:status-updates"

// OrderUpdateBus distributes order updates across replicas so an SSE client
// connected to replica B receives updates consumed by replica A.
type OrderUpdateBus interface {
	Publish(ctx context.Context, order model.Order) error
	Subscribe(ctx context.Context, fn func(model.Order)) error
}

type redisOrderBus struct {
	client *redis.Client
}

func NewRedisOrderBus(client *redis.Client) OrderUpdateBus {
	return &redisOrderBus{client: client}
}

func (b *redisOrderBus) Publish(ctx context.Context, order model.Order) error {
	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, orderUpdatesChannel, payload).Err()
}

// Subscribe starts a goroutine that forwards every update on the channel to
// fn. It returns after the subscription is confirmed, so updates published
// afterwards are guaranteed to be received. The goroutine exits when ctx is
// cancelled.
func (b *redisOrderBus) Subscribe(ctx context.Context, fn func(model.Order)) error {
	sub := b.client.Subscribe(ctx, orderUpdatesChannel)
	// Receive forces the subscription handshake before we return.
	if _, err := sub.Receive(ctx); err != nil {
		sub.Close()
		return err
	}

	go func() {
		defer sub.Close()
		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var order model.Order
				if err := json.Unmarshal([]byte(msg.Payload), &order); err != nil {
					logger.Warn("invalid order update on bus", logger.Err(err))
					continue
				}
				fn(order)
			}
		}
	}()
	return nil
}

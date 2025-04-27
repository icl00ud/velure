package publisher

import (
	"encoding/json"

	"github.com/icl00ud/publish-order-service/pkg/model"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	ch       *amqp091.Channel
	exchange string
}

func NewRabbitMQPublisher(url, exchange string) (*RabbitMQPublisher, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &RabbitMQPublisher{ch: ch, exchange: exchange}, nil
}

func (r *RabbitMQPublisher) Publish(evt model.Event) error {
	body, _ := json.Marshal(evt)
	return r.ch.Publish(r.exchange, evt.Type, false, false,
		amqp091.Publishing{ContentType: "application/json", Body: body},
	)
}

func (r *RabbitMQPublisher) Close() {
	_ = r.ch.Close()
}

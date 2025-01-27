// queue/rabbitmq.go
package queue

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/icl00ud/velure-order-service/domain"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQRepository struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   string
}

func NewRabbitMQRepo() (*RabbitMQRepository, error) {
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	user := os.Getenv("RABBITMQ_USER")
	pass := os.Getenv("RABBITMQ_PASS")
	queueName := os.Getenv("RABBITMQ_QUEUE")

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir o canal RabbitMQ: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao declarar a fila RabbitMQ: %w", err)
	}

	return &RabbitMQRepository{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (r *RabbitMQRepository) PublishOrder(order domain.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("falha ao serializar a ordem: %w", err)
	}

	err = r.channel.Publish(
		"",      // exchange
		r.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("falha ao publicar a mensagem: %w", err)
	}

	return nil
}

func (r *RabbitMQRepository) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("falha ao fechar o canal RabbitMQ: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("falha ao fechar a conexão RabbitMQ: %w", err)
	}
	return nil
}

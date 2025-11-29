package queue

import "github.com/rabbitmq/amqp091-go"

type AMQPConnection interface {
	Channel() (AMQPChannel, error)
	Close() error
}

type AMQPChannel interface {
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
	Qos(prefetchCount, prefetchSize int, global bool) error
	QueueBind(queue, key, exchange string, noWait bool, args amqp091.Table) error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error
	Close() error
}

type AMQPDialer func(url string) (AMQPConnection, error)

var amqpDial AMQPDialer = func(url string) (AMQPConnection, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return &amqpConnWrapper{conn: conn}, nil
}

type rawAMQPConnection interface {
	Channel() (*amqp091.Channel, error)
	Close() error
}

// amqpConnWrapper adapts the concrete AMQP connection to the interface, enabling stubs in tests.
type amqpConnWrapper struct {
	conn rawAMQPConnection
}

func (a *amqpConnWrapper) Channel() (AMQPChannel, error) {
	return a.conn.Channel()
}

func (a *amqpConnWrapper) Close() error {
	return a.conn.Close()
}

// SetAMQPDialer allows tests to swap the dialer and returns a restore function.
func SetAMQPDialer(dialer AMQPDialer) func() {
	prev := amqpDial
	amqpDial = dialer
	return func() { amqpDial = prev }
}

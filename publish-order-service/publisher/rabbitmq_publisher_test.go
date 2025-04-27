package publisher

import "testing"

func TestNewRabbitMQPublisher_InvalidURL(t *testing.T) {
	_, err := NewRabbitMQPublisher("amqp://invalid:1234/", "ex")
	if err == nil {
		t.Fatal("esperava erro com URL inválida, mas err==nil")
	}
}

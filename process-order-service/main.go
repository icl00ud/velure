package main

import (
	"log"

	"github.com/icl00ud/process-order-service/handler"
	"github.com/icl00ud/process-order-service/queue"
	"github.com/icl00ud/process-order-service/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	paymentStorage := storage.NewPaymentStorage()
	defer paymentStorage.DB.Close()

	consumer := queue.NewConsumer()
	defer consumer.Close()

	orderConsumer := handler.NewOrderConsumer(consumer)
	orderConsumer.StartConsuming(paymentStorage)
}

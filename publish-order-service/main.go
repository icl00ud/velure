// cmd/publisher/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/icl00ud/velure-order-service/handler"
	"github.com/icl00ud/velure-order-service/queue"
	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.New()

	rabbitRepo, err := queue.NewRabbitMQRepo()
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer func() {
		if err := rabbitRepo.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}()

	orderHandler := handler.NewOrderHandler(rabbitRepo)

	app.POST("/publish-order", orderHandler.CreateOrder)

	go func() {
		log.Println("Starting the Publisher Service...")
		app.Run()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Received signal %s. Shutting down the Publisher Service...", sig)

	err = app.Shutdown(gofr.Context{})
	if err != nil {
		log.Fatalf("Error shutting down the Publisher Service: %v", err)
	}

	log.Println("Publisher Service successfully shut down.")
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/icl00ud/publish-order-service/handlers"
	"github.com/icl00ud/publish-order-service/middleware"
	"github.com/icl00ud/publish-order-service/queue"
	"github.com/icl00ud/publish-order-service/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env found! Using the system environment variables.")
	}

	rabbitRepo := queue.NewRabbitMQRepo()
	defer rabbitRepo.Close()

	dbStorage := storage.NewStorage()
	defer dbStorage.DB.Close()

	orderHandler := handlers.NewOrderHandler(dbStorage, rabbitRepo)

	mux := http.NewServeMux()

	loggedMux := middleware.LoggingMiddleware(mux)

	mux.HandleFunc("/create-order", orderHandler.CreateOrder)

	port := os.Getenv("PUBLISH_ORDER_SERVICE_APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Publisher Service initialized at PORT: %s...", port)
	if err := http.ListenAndServe(":"+port, loggedMux); err != nil {
		log.Fatalf("Failed to open server: %v", err)
	}
}

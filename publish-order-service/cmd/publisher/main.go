package main

import (
	"log"
	"net/http"
	"os"

	"github.com/icl00ud/publish-order-service/client"
	"github.com/icl00ud/publish-order-service/handlers"
	"github.com/icl00ud/publish-order-service/middleware"
	"github.com/icl00ud/publish-order-service/queue"
	"github.com/icl00ud/publish-order-service/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Nenhum arquivo .env encontrado. Usando variáveis de ambiente existentes.")
	}

	rabbitRepo, err := queue.NewRabbitMQRepo()
	if err != nil {
		log.Fatalf("Erro ao conectar ao RabbitMQ: %v", err)
	}
	defer func() {
		if err := rabbitRepo.Close(); err != nil {
			log.Printf("Erro ao fechar a conexão RabbitMQ: %v", err)
		}
	}()

	dbStorage, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer dbStorage.DB.Close()

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		log.Fatalf("PRODUCT_SERVICE_URL não está definido no .env")
	}
	productClient := client.NewProductClient(productServiceURL)

	orderHandler := handlers.NewOrderHandler(productClient, dbStorage, rabbitRepo)

	mux := http.NewServeMux()

	loggedMux := middleware.LoggingMiddleware(mux)

	mux.HandleFunc("/create-order", orderHandler.CreateOrder)

	port := os.Getenv("PUBLISH_ORDER_SERVICE_APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Publisher Service iniciado na porta %s...", port)
	if err := http.ListenAndServe(":"+port, loggedMux); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

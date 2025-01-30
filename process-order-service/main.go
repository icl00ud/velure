package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/icl00ud/process-order-service/handler"
	"github.com/icl00ud/process-order-service/queue"
)

func main() {
	// Inicializa o repositório RabbitMQ
	rabbitRepo, err := queue.NewRabbitMQRepo()
	if err != nil {
		log.Fatalf("Erro ao conectar ao RabbitMQ: %v", err)
	}
	defer func() {
		if err := rabbitRepo.Close(); err != nil {
			log.Printf("Erro ao fechar a conexão RabbitMQ: %v", err)
		}
	}()

	// Inicializa o consumidor de pedidos
	orderConsumer := handler.NewOrderConsumer(rabbitRepo)
	orderConsumer.StartConsuming()

	log.Println("Consumer Service iniciado e escutando por eventos...")

	// Escuta por sinais de interrupção para finalizar a aplicação
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Recebido sinal %s. Encerrando o Consumer Service...", sig)

	// Encerramento limpo
	log.Println("Consumer Service encerrado com sucesso.")
}

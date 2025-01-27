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
		log.Fatalf("Erro ao conectar ao RabbitMQ: %v", err)
	}
	defer func() {
		if err := rabbitRepo.Close(); err != nil {
			log.Printf("Erro ao fechar a conexão RabbitMQ: %v", err)
		}
	}()

	orderHandler := handler.NewOrderHandler(rabbitRepo)

	app.POST("/publish-order", orderHandler.CreateOrder)

	go func() {
		log.Println("Iniciando a aplicação...")
		app.Run()
		log.Println("Aplicação iniciada!")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Recebido sinal %s. Encerrando a aplicação...", sig)

	err = app.Shutdown(gofr.Context{})
	if err != nil {
		log.Fatalf("Erro ao encerrar a aplicação: %v", err)
	}

	if err := rabbitRepo.Close(); err != nil {
		log.Printf("Erro ao fechar a conexão RabbitMQ: %v", err)
	}

	log.Println("Aplicação encerrada com sucesso.")
}

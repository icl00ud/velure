package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/icl00ud/publish-order-service/domain"
)

type Storage struct {
	DB *sql.DB
}

func NewStorage() (*Storage, error) {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_HOST_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DATABASE_NAME")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to the database successfully")

	// Executa as migrações
	migrationPath := "file://migrations"
	m, err := migrate.New(migrationPath, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply.")
	} else {
		log.Println("Migrations applied successfully.")
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) CreateOrder(order domain.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal cart items: %w", err)
	}

	query := `
		INSERT INTO orders (id, items, total, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`
	_, err = s.DB.Exec(query, order.ID, itemsJSON, order.Total, order.Status, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}
	return nil
}

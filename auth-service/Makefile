build: ## Build the Go application
	go build -o bin/auth-service .

run: ## Run the application
	go run .

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

docker-build: ## Build Docker image
	docker build -t velure-auth-service .

docker-run: ## Run Docker container
	docker run -p 3020:3020 --env-file .env velure-auth-service

migrate-up: ## Run database migrations up
	migrate -path ./migrations -database "$(POSTGRES_URL)" up

migrate-down: ## Run database migrations down
	migrate -path ./migrations -database "$(POSTGRES_URL)" down

migrate-create: ## Create a new migration file
	migrate create -ext sql -dir ./migrations -seq $(name)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

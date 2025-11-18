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

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -benchtime=10s ./internal/services/...

load-test: ## Run load tests with k6
	@echo "Starting load tests..."
	k6 run --vus 100 --duration 30s k6/performance-test.js

performance-test: ## Run comprehensive performance tests
	@echo "Running comprehensive performance tests..."
	k6 run k6/performance-test.js

perf-custom: ## Run custom performance test (VUS=100 DURATION=30s)
	@echo "Running custom performance test..."
	k6 run --vus $(VUS) --duration $(DURATION) k6/performance-test.js

spike-test: ## Run spike test
	@echo "Running spike test..."
	k6 run --vus 500 --duration 1m k6/performance-test.js

stress-test: ## Run stress test
	@echo "Running stress test..."
	k6 run --vus 1000 --duration 5m k6/performance-test.js

profile-cpu: ## Profile CPU usage
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./internal/services/...
	go tool pprof -http=:8080 cpu.prof

profile-mem: ## Profile memory usage
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./internal/services/...
	go tool pprof -http=:8080 mem.prof

docker-run-optimized: ## Run Docker container with performance tuning
	docker run -p 3020:3020 \
		-e BCRYPT_WORKERS=20 \
		-e ENABLE_TOKEN_CACHE=true \
		-e TOKEN_CACHE_TTL=300 \
		-e BCRYPT_COST=10 \
		velure-auth-service

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

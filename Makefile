# Velure - Cloud-Native E-Commerce Platform
# Developer shortcuts for local, cloud, docs, and validation workflows.

GO_MODULES := $(shell find services -mindepth 2 -maxdepth 2 -name go.mod -exec dirname {} \; | sort)
UI_DIR := services/ui-service

.PHONY: help local-up local-down local-dev local-dev-down cloud-up cloud-down cloud-urls test lint check

# Default target
help: ## Show available commands
	@echo "Velure - Cloud-Native E-Commerce Platform"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start:"
	@echo "  make local-up     # Start the local stack"
	@echo "  make check        # Run validation checks"
	@echo ""

# =============================================================================
# Local Development
# =============================================================================

local-up: ## Start the full local stack (infrastructure, services, monitoring)
	@echo "Starting the full local environment..."
	@echo ""
	@echo "Starting infrastructure, services, and monitoring..."
	cd infrastructure/local && docker-compose -f docker-compose.yaml up -d
	cd infrastructure/local && docker-compose -f docker-compose.monitoring.yaml up -d
	@echo ""
	@echo "Waiting 20 seconds for startup..."
	@sleep 20
	@echo ""
	@echo "Local environment is ready."
	@echo ""
	@echo "Access URLs:"
	@echo ""
	@echo "Application:  http://localhost"
	@echo "Grafana:      http://localhost:3000 (admin/admin)"
	@echo "Prometheus:   http://localhost:9090"
	@echo "RabbitMQ:     http://localhost:15672 (admin/admin_password)"
	@echo "cAdvisor:     http://localhost:8080"
	@echo ""
	@echo "Container status:"
	@docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(auth|product|publish|process|ui-service|postgres|mongodb|redis|rabbitmq|caddy|grafana|prometheus)" || true
	@echo ""
	@echo "Stop the stack with: make local-down"
	@echo ""

local-down: ## Stop the local stack and remove containers and volumes
	@echo "Stopping the local environment..."
	@echo ""
	@echo "Stopping containers..."
	cd infrastructure/local && docker-compose -f docker-compose.monitoring.yaml down -v --remove-orphans || true
	cd infrastructure/local && docker-compose -f docker-compose.yaml down -v --remove-orphans
	@echo ""
	@echo "Pruning unused Docker resources..."
	docker system prune -f --volumes
	@echo ""
	@echo "Removing Docker networks..."
	docker network rm local_auth 2>/dev/null || true
	docker network rm local_order 2>/dev/null || true
	docker network rm local_frontend 2>/dev/null || true
	@echo ""
	@echo "Local environment removed."
	@echo ""

local-dev: ## Start stack with hot-reload (Air for Go services, Vite HMR for UI)
	@echo "Starting local dev environment with hot-reload..."
	@echo ""
	@echo "Building dev images and starting services..."
	@echo "Note: First run builds dev images (~1-2 min). Subsequent starts are faster."
	cd infrastructure/local && docker-compose \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		up --build -d
	@echo ""
	@echo "Seeding MongoDB (idempotent — skips if products exist)..."
	@bash -c 'for i in $$(seq 1 30); do \
		docker exec mongodb mongosh --quiet -u velure_user -p velure_password --authenticationDatabase admin --eval "db.runCommand({ping:1})" >/dev/null 2>&1 && break; \
		sleep 1; \
	done'
	@docker cp services/product-service/mongo-init.js mongodb:/tmp/seed.js
	@docker exec mongodb mongosh --quiet -u velure_user -p velure_password --authenticationDatabase admin --eval "load('/tmp/seed.js')" | tail -5 || echo "Seed failed (non-fatal)"
	@echo ""
	@echo "Dev environment is ready."
	@echo ""
	@echo "Access URLs:"
	@echo ""
	@echo "Application:  http://localhost"
	@echo "RabbitMQ:     http://localhost:15672 (admin/admin_password)"
	@echo ""
	@echo "Hot-reload:"
	@echo "  Go services  — Air rebuilds on .go file change (~1s)"
	@echo "  UI           — Vite HMR updates browser instantly on .tsx/.ts change"
	@echo ""
	@echo "Watch logs:  docker logs -f <container-name>"
	@echo "Stop:        make local-dev-down"
	@echo ""

local-dev-down: ## Stop the hot-reload dev stack and remove all volumes (including ui_node_modules)
	@echo "Stopping the dev environment..."
	@echo ""
	cd infrastructure/local && docker-compose \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		down -v --remove-orphans
	@echo ""
	@echo "Pruning unused Docker resources..."
	docker system prune -f --volumes
	@echo ""
	@echo "Removing Docker networks..."
	docker network rm local_auth 2>/dev/null || true
	docker network rm local_order 2>/dev/null || true
	docker network rm local_frontend 2>/dev/null || true
	@echo ""
	@echo "Dev environment removed."
	@echo ""

# =============================================================================
# CLOUD (AWS EKS)
# =============================================================================

cloud-up: ## Provision the full AWS stack (Terraform, Kubernetes, monitoring)
	@echo "Starting full AWS deployment..."
	@echo ""
	@echo "Phase 1: Provision AWS infrastructure with Terraform"
	@echo ""
	@echo "Resources to be created:"
	@echo "  - VPC and public/private subnets across 2 AZs"
	@echo "  - EKS cluster and node groups (t3.medium)"
	@echo "  - Two RDS PostgreSQL instances (auth and orders)"
	@echo "  - AmazonMQ (RabbitMQ)"
	@echo "  - Route53 hosted zone"
	@echo "  - Secrets Manager entries"
	@echo ""
	@echo "Estimated time: about 15 minutes"
	@echo ""
	cd infrastructure/terraform && terraform init -upgrade
	cd infrastructure/terraform && terraform apply -auto-approve
	@echo ""
	@echo "AWS infrastructure created."
	@echo ""
	@echo "Phase 2: Configure Kubernetes with deploy-eks.sh"
	@echo ""
	@echo "Components to be installed:"
	@echo "  - AWS Load Balancer Controller"
	@echo "  - Metrics Server and External Secrets Operator"
	@echo "  - Datastores (MongoDB, Redis, RabbitMQ)"
	@echo "  - Monitoring stack (Prometheus and Grafana)"
	@echo "  - Velure services (auth, product, orders, UI)"
	@echo ""
	@echo "Estimated time: about 10 minutes"
	@echo ""
	chmod +x infrastructure/scripts/deploy-eks.sh
	./infrastructure/scripts/deploy-eks.sh
	@echo ""
	@echo "Cloud deployment completed."
	@echo ""
	@echo "To retrieve access URLs, run:"
	@echo "  make cloud-urls"
	@echo ""

cloud-down: ## Destroy all AWS infrastructure and force-delete Velure secrets
	@echo "WARNING: This action is destructive and irreversible."
	@echo ""
	@echo "This will remove:"
	@echo "  - All Terraform-managed infrastructure (EKS, RDS, VPC, etc.)"
	@echo "  - Velure Secrets Manager entries, force-deleted without recovery"
	@echo ""
	@printf "Type 'DESTROY' to confirm: "; \
	read confirm; \
	if [ "$$confirm" != "DESTROY" ]; then \
		echo "Canceled."; \
		exit 1; \
	fi
	@echo ""
	@echo "Phase 1: Destroy Terraform infrastructure..."
	@echo ""
	cd infrastructure/terraform && terraform destroy -auto-approve
	@echo ""
	@echo "Terraform destroy completed."
	@echo ""
	@echo "Phase 2: Force-delete Velure secrets..."
	@echo ""
	@aws secretsmanager list-secrets --region us-east-1 --query 'SecretList[?starts_with(Name, `velure-`)].Name' --output text | \
	tr '\t' '\n' | while read secret; do \
		if [ -n "$$secret" ]; then \
			echo "  Deleting $$secret..."; \
			aws secretsmanager delete-secret --secret-id "$$secret" --force-delete-without-recovery --region us-east-1 2>/dev/null || true; \
		fi; \
	done
	@echo "Secrets deleted."
	@echo ""
	@echo "AWS infrastructure removed."
	@echo ""

cloud-urls: ## Show AWS access URLs for the application and observability tools
	@echo "AWS access URLs"
	@echo ""
	@echo "Frontend (UI):"
	@UI_URL=$$(kubectl get ingress velure-ui -n frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null); \
	if [ -n "$$UI_URL" ]; then \
		echo "   http://$$UI_URL"; \
	else \
		echo "   Not available yet; the ALB may still be provisioning."; \
		echo "   Run this target again in a few minutes."; \
	fi
	@echo ""
	@echo "Grafana (observability):"
	@GRAFANA_URL=$$(kubectl get ingress grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null); \
	if [ -n "$$GRAFANA_URL" ]; then \
		echo "   http://$$GRAFANA_URL"; \
		echo "   Credentials: admin / admin"; \
	else \
		echo "   Not exposed through Ingress."; \
		echo "   Use port-forward:"; \
		echo "   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"; \
		echo "   Then open: http://localhost:3000 (admin/admin)"; \
	fi
	@echo ""
	@echo "Prometheus:"
	@echo "   kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090"
	@echo "   Then open: http://localhost:9090"
	@echo ""
	@echo "Ingress status:"
	@kubectl get ingress -A 2>/dev/null || echo "  Unable to query ingresses. Is kubectl configured?"
	@echo ""

# -----------------------------------------------------------------------------
# Validation
# -----------------------------------------------------------------------------

test: ## Run Go service tests and UI tests when available
	@echo "Running Go service tests..."
	@set -e; \
	for dir in $(GO_MODULES); do \
		echo "==> $$dir"; \
		(cd "$$dir" && go test ./...); \
	done
	@if [ -f "$(UI_DIR)/package.json" ]; then \
		echo "Running UI tests..."; \
		(cd "$(UI_DIR)" && npm run test:run --if-present); \
	else \
		echo "Skipping UI tests; $(UI_DIR)/package.json was not found."; \
	fi

check: ## Run formatting, vet, and tests
	@echo "Checking Go formatting..."
	@unformatted=$$(find services -name '*.go' -exec gofmt -l {} +); \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted"; \
		echo "Go files need formatting. Run gofmt on the files above."; \
		exit 1; \
	fi
	@echo "Running Go vet..."
	@set -e; \
	for dir in $(GO_MODULES); do \
		echo "==> $$dir"; \
		(cd "$$dir" && go vet ./...); \
	done
	@$(MAKE) test

lint: ## Run lint checks (Go vet + UI lint)
	@echo "Running Go vet..."
	@set -e; \
	for dir in $(GO_MODULES); do \
		echo "==> $$dir"; \
		(cd "$$dir" && go vet ./...); \
	done
	@if [ -f "$(UI_DIR)/package.json" ]; then \
		echo "Running UI lint..."; \
		(cd "$(UI_DIR)" && npm run lint --if-present); \
	else \
		echo "Skipping UI lint; $(UI_DIR)/package.json was not found."; \
	fi

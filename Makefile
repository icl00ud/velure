# Velure - Microservices Platform Makefile
# Automação para desenvolvimento, testes e deployment

.PHONY: help dev build test clean deploy-local deploy-k8s deploy-aws

# Default target
help: ## Mostrar esta mensagem de ajuda
	@echo "Velure - Comandos disponíveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# DESENVOLVIMENTO LOCAL
# =============================================================================

dev: ## Iniciar ambiente de desenvolvimento completo
	@echo "🚀 Iniciando ambiente de desenvolvimento..."
	cd infrastructure/local && docker-compose up -d
	@echo "✅ Dependências iniciadas. Execute 'make dev-services' para subir os serviços."

dev-services: ## Subir todos os serviços em paralelo (desenvolvimento)
	@echo "🔧 Iniciando todos os serviços..."
	@trap 'kill 0' INT; \
	cd services/auth-service && go run main.go & \
	cd services/product-service && go run main.go & \
	cd services/publish-order-service && go run main.go & \
	cd services/process-order-service && go run main.go & \
	cd services/ui-service && npm start & \
	wait

dev-stop: ## Parar ambiente de desenvolvimento
	@echo "🛑 Parando ambiente de desenvolvimento..."
	cd infrastructure/local && docker-compose down
	@echo "✅ Ambiente parado."

dev-clean: ## Limpar volumes e dados do desenvolvimento
	@echo "🧹 Limpando dados de desenvolvimento..."
	cd infrastructure/local && docker-compose down -v --remove-orphans
	docker system prune -f
	@echo "✅ Limpeza concluída."

# =============================================================================
# BUILD E TESTES
# =============================================================================

build: ## Build de todos os serviços
	@echo "🔨 Building todos os serviços..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		echo "Building $$service..."; \
		cd services/$$service && go build -o bin/$$service main.go && cd ../..; \
	done
	@echo "Building ui-service..."
	cd services/ui-service && npm run build
	@echo "✅ Build concluído."

test: ## Executar todos os testes
	@echo "🧪 Executando testes..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		echo "Testing $$service..."; \
		cd services/$$service && go test ./... && cd ../..; \
	done
	cd services/ui-service && npm test
	@echo "✅ Testes concluídos."

test-load: ## Executar testes de carga (k6)
	@echo "📊 Executando testes de carga..."
	cd tests/load && ./run-all-tests.sh
	@echo "✅ Testes de carga concluídos."

clean: ## Limpar binários e cache
	@echo "🧹 Limpando binários e cache..."
	find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name "node_modules" -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.log" -type f -delete 2>/dev/null || true
	go clean -cache -modcache -testcache
	@echo "✅ Limpeza concluída."

# =============================================================================
# QUALIDADE DE CÓDIGO
# =============================================================================

lint: ## Executar linting em todos os serviços
	@echo "🔍 Executando linting..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		echo "Linting $$service..."; \
		cd services/$$service && golangci-lint run && cd ../..; \
	done
	cd services/ui-service && npm run lint
	@echo "✅ Linting concluído."

format: ## Formatar código
	@echo "💅 Formatando código..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		cd services/$$service && go fmt ./... && cd ../..; \
	done
	cd services/ui-service && npm run format
	@echo "✅ Formatação concluída."

security: ## Verificar vulnerabilidades de segurança
	@echo "🔒 Verificando segurança..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		echo "Security scan $$service..."; \
		cd services/$$service && gosec ./... && cd ../..; \
	done
	cd services/ui-service && npm audit
	@echo "✅ Verificação de segurança concluída."

# =============================================================================
# DOCKER
# =============================================================================

docker-build: ## Build de todas as imagens Docker
	@echo "🐳 Building imagens Docker..."
	@for service in auth-service product-service publish-order-service process-order-service ui-service; do \
		echo "Building docker image for $$service..."; \
		cd services/$$service && docker build -t velure/$$service:latest . && cd ../..; \
	done
	@echo "✅ Imagens Docker criadas."

docker-push: ## Push das imagens para registry
	@echo "📤 Pushing imagens para registry..."
	@for service in auth-service product-service publish-order-service process-order-service ui-service; do \
		docker push velure/$$service:latest; \
	done
	@echo "✅ Push concluído."

# =============================================================================
# KUBERNETES LOCAL
# =============================================================================

k8s-setup: ## Configurar Kubernetes local (namespaces, secrets)
	@echo "☸️ Configurando Kubernetes local..."
	kubectl create namespace database || true
	kubectl create namespace order || true
	kubectl create namespace authentication || true
	kubectl create namespace frontend || true
	@echo "✅ Namespaces criados."

k8s-deploy-infra: ## Deploy da infraestrutura (bancos, cache, filas)
	@echo "☸️ Deploying infraestrutura..."
	helm upgrade --install postgres infrastructure/kubernetes/charts/postgresql -n database
	helm upgrade --install mongodb infrastructure/kubernetes/charts/mongodb -n database
	helm upgrade --install redis infrastructure/kubernetes/charts/redis -n database
	helm upgrade --install rabbitmq infrastructure/kubernetes/charts/velure-rabbitmq -n order
	@echo "✅ Infraestrutura deployada."

k8s-deploy-services: ## Deploy dos microserviços
	@echo "☸️ Deploying serviços..."
	helm upgrade --install velure-auth infrastructure/kubernetes/charts/velure-auth -n authentication
	helm upgrade --install velure-product infrastructure/kubernetes/charts/velure-product -n order
	helm upgrade --install velure-publish-order infrastructure/kubernetes/charts/velure-publish-order -n order
	helm upgrade --install velure-process-order infrastructure/kubernetes/charts/velure-process-order -n order
	helm upgrade --install velure-ui infrastructure/kubernetes/charts/velure-ui -n frontend
	@echo "✅ Serviços deployados."

k8s-deploy: k8s-setup k8s-deploy-infra k8s-deploy-services ## Deploy completo no Kubernetes local

k8s-destroy: ## Remover tudo do Kubernetes local
	@echo "🗑️ Removendo deployment Kubernetes..."
	helm uninstall velure-ui -n frontend || true
	helm uninstall velure-auth -n authentication || true
	helm uninstall velure-product -n order || true
	helm uninstall velure-publish-order -n order || true
	helm uninstall velure-process-order -n order || true
	helm uninstall rabbitmq -n order || true
	helm uninstall postgres -n database || true
	helm uninstall mongodb -n database || true
	helm uninstall redis -n database || true
	kubectl delete namespace frontend || true
	kubectl delete namespace authentication || true
	kubectl delete namespace order || true
	kubectl delete namespace database || true
	@echo "✅ Kubernetes limpo."

k8s-status: ## Verificar status dos pods
	@echo "📊 Status dos pods:"
	kubectl get pods -A | grep velure

# =============================================================================
# AWS EKS
# =============================================================================

aws-plan: ## Planejar infraestrutura AWS
	@echo "☁️ Planejando infraestrutura AWS..."
	cd infrastructure/terraform && terraform plan
	@echo "✅ Plano gerado."

aws-deploy: ## Deploy da infraestrutura AWS
	@echo "☁️ Deploying infraestrutura AWS..."
	cd infrastructure/terraform && terraform apply -auto-approve
	@echo "✅ Infraestrutura AWS criada."

aws-destroy: ## Destruir infraestrutura AWS
	@echo "⚠️ CUIDADO: Isso deletará TODA a infraestrutura AWS!"
	@read -p "Digite 'CONFIRM' para continuar: " confirm && [ "$$confirm" = "CONFIRM" ]
	cd infrastructure/terraform && terraform destroy -auto-approve
	@echo "✅ Infraestrutura AWS destruída."

aws-status: ## Verificar status do cluster EKS
	@echo "📊 Status do cluster EKS:"
	aws eks describe-cluster --name velure-prod --region us-east-1 --query 'cluster.status'
	kubectl get nodes

aws-kubeconfig: ## Configurar kubectl para EKS
	@echo "⚙️ Configurando kubectl para EKS..."
	aws eks update-kubeconfig --region us-east-1 --name velure-prod
	@echo "✅ kubectl configurado."

# =============================================================================
# MONITORAMENTO
# =============================================================================

monitoring-setup: ## Configurar stack de monitoramento
	@echo "📊 Configurando monitoramento..."
	cd tools/monitoring && docker-compose up -d
	@echo "✅ Prometheus e Grafana disponíveis:"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Grafana: http://localhost:3000 (admin/admin)"

monitoring-stop: ## Parar stack de monitoramento
	@echo "🛑 Parando monitoramento..."
	cd tools/monitoring && docker-compose down
	@echo "✅ Monitoramento parado."

logs: ## Verificar logs dos serviços (Kubernetes)
	@echo "📋 Logs dos serviços:"
	@for ns in authentication order frontend; do \
		echo "=== Namespace: $$ns ==="; \
		kubectl logs -n $$ns -l app.kubernetes.io/instance=velure --tail=10; \
	done

health: ## Verificar health dos serviços
	@echo "🏥 Verificando health dos serviços..."
	@for port in 3020 3010 3030 3040; do \
		echo -n "Serviço na porta $$port: "; \
		curl -s -o /dev/null -w "%{http_code}" http://localhost:$$port/health || echo "❌ Indisponível"; \
		echo ""; \
	done

# =============================================================================
# UTILITÁRIOS
# =============================================================================

deps: ## Instalar dependências de todos os serviços
	@echo "📦 Instalando dependências..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		echo "Installing deps for $$service..."; \
		cd services/$$service && go mod download && cd ../..; \
	done
	cd services/ui-service && npm install
	@echo "✅ Dependências instaladas."

update-deps: ## Atualizar dependências
	@echo "📦 Atualizando dependências..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		cd services/$$service && go get -u ./... && go mod tidy && cd ../..; \
	done
	cd services/ui-service && npm update
	@echo "✅ Dependências atualizadas."

docs: ## Servir documentação local
	@echo "📚 Servindo documentação..."
	cd docs && python3 -m http.server 8080
	@echo "📖 Documentação disponível em: http://localhost:8080"

version: ## Mostrar versões das ferramentas
	@echo "🔧 Versões das ferramentas:"
	@echo "Go: $$(go version)"
	@echo "Node: $$(node --version)"
	@echo "Docker: $$(docker --version)"
	@echo "Kubernetes: $$(kubectl version --client --short)"
	@echo "Helm: $$(helm version --short)"
	@echo "Terraform: $$(terraform version | head -1)"

# =============================================================================
# DEVELOPMENT HELPERS
# =============================================================================

new-service: ## Criar template de novo serviço (make new-service NAME=my-service)
	@if [ -z "$(NAME)" ]; then echo "❌ Use: make new-service NAME=nome-do-servico"; exit 1; fi
	@echo "🆕 Criando novo serviço: $(NAME)"
	mkdir -p services/$(NAME)
	cd services/$(NAME) && go mod init $(NAME)
	@echo "✅ Serviço $(NAME) criado em services/$(NAME)"

generate-docs: ## Gerar documentação da API (Swagger)
	@echo "📝 Gerando documentação da API..."
	@for service in auth-service product-service publish-order-service process-order-service; do \
		cd services/$$service && swag init && cd ../..; \
	done
	@echo "✅ Documentação gerada."

backup-local: ## Backup dos dados locais
	@echo "💾 Fazendo backup dos dados locais..."
	mkdir -p backups
	cd infrastructure/local && docker-compose exec postgres pg_dump -U user > ../../backups/postgres-$$(date +%Y%m%d).sql
	@echo "✅ Backup salvo em backups/"

# =============================================================================
# ALIASES ÚTEIS
# =============================================================================

start: dev ## Alias para 'make dev'
stop: dev-stop ## Alias para 'make dev-stop'
restart: dev-stop dev ## Reiniciar ambiente de desenvolvimento
deploy: k8s-deploy ## Alias para 'make k8s-deploy'
destroy: k8s-destroy ## Alias para 'make k8s-destroy'
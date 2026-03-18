# Velure - Cloud-Native E-Commerce Platform
# Simplified Makefile with essential commands only

.PHONY: help local-up local-down cloud-up cloud-down cloud-urls

# Default target
help: ## Mostrar comandos disponíveis
	@echo "╦  ╦┌─┐┬  ┬ ┬┬─┐┌─┐"
	@echo "╚╗╔╝├┤ │  │ │├┬┘├┤ "
	@echo " ╚╝ └─┘┴─┘└─┘┴└─└─┘"
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "                    COMANDOS ESSENCIAIS                        "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start:"
	@echo "  make local-up    # Desenvolvimento local"
	@echo "  make cloud-up    # Deploy completo AWS"
	@echo ""

# =============================================================================
# DESENVOLVIMENTO LOCAL
# =============================================================================

local-up: ## Subir aplicação COMPLETA localmente (infra + services + monitoring)
	@echo "🚀 Iniciando ambiente LOCAL completo..."
	@echo ""
	@echo "📦 Criando redes Docker..."
	@docker network create local_auth 2>/dev/null || echo "  ✓ Rede local_auth já existe"
	@docker network create local_order 2>/dev/null || echo "  ✓ Rede local_order já existe"
	@docker network create local_frontend 2>/dev/null || echo "  ✓ Rede local_frontend já existe"
	@echo ""
	@echo "📦 Subindo infraestrutura + serviços + monitoramento..."
	cd infrastructure/local && docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml up -d
	@echo ""
	@echo "⏳ Aguardando inicialização (20 segundos)..."
	@sleep 20
	@echo ""
	@echo "✅ AMBIENTE LOCAL PRONTO!"
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "                        ACESSOS                                "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "🌐 Aplicação:     https://velure.local"
	@echo "📊 Grafana:       http://localhost:3000 (admin/admin)"
	@echo "📈 Prometheus:    http://localhost:9090"
	@echo "🐰 RabbitMQ:      http://localhost:15672 (admin/admin_password)"
	@echo "📦 cAdvisor:      http://localhost:8080"
	@echo ""
	@echo "📋 Status:"
	@docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(auth|product|publish|process|ui-service|postgres|mongodb|redis|rabbitmq|caddy|grafana|prometheus)" || true
	@echo ""
	@echo "💡 Para derrubar: make local-down"
	@echo ""

local-down: ## Derrubar aplicação local completa (remove containers + volumes)
	@echo "🛑 Derrubando ambiente LOCAL..."
	@echo ""
	@echo "Parando containers..."
	cd infrastructure/local && docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml down -v --remove-orphans
	@echo ""
	@echo "Limpando recursos órfãos..."
	docker system prune -f --volumes
	@echo ""
	@echo "Removendo redes..."
	docker network rm local_auth 2>/dev/null || true
	docker network rm local_order 2>/dev/null || true
	docker network rm local_frontend 2>/dev/null || true
	@echo ""
	@echo "✅ AMBIENTE LOCAL REMOVIDO!"
	@echo ""

# =============================================================================
# CLOUD (AWS EKS)
# =============================================================================

cloud-up: ## Subir infraestrutura COMPLETA na AWS (Terraform + Kubernetes + Monitoring)
	@echo "☁️  Iniciando deployment COMPLETO na AWS..."
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "  FASE 1: Provisionando infraestrutura AWS (Terraform)         "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Recursos que serão criados:"
	@echo "  • VPC + Subnets (public/private em 2 AZs)"
	@echo "  • EKS Cluster + Node Groups (t3.medium)"
	@echo "  • RDS PostgreSQL x2 (auth + orders)"
	@echo "  • AmazonMQ (RabbitMQ)"
	@echo "  • Route53 Hosted Zone"
	@echo "  • Secrets Manager"
	@echo ""
	@echo "⏳ Tempo estimado: ~15 minutos"
	@echo ""
	cd infrastructure/terraform && terraform init -upgrade
	cd infrastructure/terraform && terraform apply -auto-approve
	@echo ""
	@echo "✅ Infraestrutura AWS criada!"
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "  FASE 2: Configurando Kubernetes (deploy-eks.sh)              "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Componentes que serão instalados:"
	@echo "  • AWS Load Balancer Controller"
	@echo "  • Metrics Server + External Secrets Operator"
	@echo "  • Datastores (MongoDB, Redis, RabbitMQ)"
	@echo "  • Monitoring Stack (Prometheus + Grafana)"
	@echo "  • Velure Services (auth, product, orders, UI)"
	@echo ""
	@echo "⏳ Tempo estimado: ~10 minutos"
	@echo ""
	chmod +x scripts/deploy-eks.sh
	./scripts/deploy-eks.sh
	@echo ""
	@echo "✅ DEPLOYMENT CLOUD COMPLETO!"
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "Para obter URLs de acesso, execute:"
	@echo "  make cloud-urls"
	@echo ""

cloud-down: ## Destruir TODA infraestrutura AWS + deletar secrets forçadamente
	@echo "⚠️  ATENÇÃO: Esta ação é DESTRUTIVA e IRREVERSÍVEL!"
	@echo ""
	@echo "Será removido:"
	@echo "  • Toda a infraestrutura gerenciada pelo Terraform (EKS, RDS, VPC etc.)"
	@echo "  • Secrets Manager (FORÇADO - mesmo pendentes de deleção)"
	@echo ""
	@read -p "Digite 'DESTROY' para confirmar: " confirm; \
	if [ "$$confirm" != "DESTROY" ]; then \
		echo "❌ Cancelado."; \
		exit 1; \
	fi
	@echo ""
	@echo "🗑️  Fase 1: Destruindo infraestrutura Terraform..."
	@echo ""
	cd infrastructure/terraform && terraform destroy -auto-approve
	@echo ""
	@echo "✅ Terraform destroy concluído."
	@echo ""
	@echo "🗑️  Fase 2: Deletando secrets forçadamente..."
	@echo ""
	@aws secretsmanager list-secrets --region us-east-1 --query 'SecretList[?starts_with(Name, `velure-`)].Name' --output text | \
	tr '\t' '\n' | while read secret; do \
		if [ -n "$$secret" ]; then \
			echo "  Deletando $$secret..."; \
			aws secretsmanager delete-secret --secret-id "$$secret" --force-delete-without-recovery --region us-east-1 2>/dev/null || true; \
		fi; \
	done
	@echo "✅ Secrets deletados."
	@echo ""
	@echo "✅ INFRAESTRUTURA AWS COMPLETAMENTE REMOVIDA!"
	@echo ""

cloud-urls: ## Mostrar URLs de acesso da aplicação na AWS
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "                    URLs DE ACESSO (AWS)                       "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "🌐 Frontend (UI):"
	@UI_URL=$$(kubectl get ingress velure-ui -n frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null); \
	if [ -n "$$UI_URL" ]; then \
		echo "   http://$$UI_URL"; \
	else \
		echo "   ⏳ Ainda não disponível (ALB sendo criado)"; \
		echo "   Execute novamente em alguns minutos"; \
	fi
	@echo ""
	@echo "📊 Grafana (Observabilidade):"
	@GRAFANA_URL=$$(kubectl get ingress grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null); \
	if [ -n "$$GRAFANA_URL" ]; then \
		echo "   http://$$GRAFANA_URL"; \
		echo "   Credenciais: admin / admin"; \
	else \
		echo "   ⏳ Não exposto via Ingress"; \
		echo "   Use port-forward:"; \
		echo "   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"; \
		echo "   Depois acesse: http://localhost:3000 (admin/admin)"; \
	fi
	@echo ""
	@echo "📈 Prometheus:"
	@echo "   kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090"
	@echo "   Depois acesse: http://localhost:9090"
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Status dos Ingresses:"
	@kubectl get ingress -A 2>/dev/null || echo "  ⚠️  Erro ao buscar ingresses (kubectl configurado?)"
	@echo ""

# -----------------------------------------------------------------------------
# Documentação
# -----------------------------------------------------------------------------

docs-install: ## Instalar dependências da documentação
	@echo "📦 Instalando dependências da documentação..."
	cd docs-site && npm install

docs-up: ## Subir o portal de documentação localmente
	@echo "📚 Iniciando portal de documentação na porta 3000..."
	cd docs-site && npm run start

docs-build: ## Buildar o portal de documentação para produção
	@echo "🏗️  Buildando portal de documentação..."
	cd docs-site && npm run build

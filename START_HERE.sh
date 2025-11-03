#!/bin/bash

# Velure - Script de Inicialização Rápida
# Este script facilita o início do projeto com monitoramento

set -e

BOLD="\033[1m"
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

echo ""
echo -e "${BOLD}${BLUE}================================================${NC}"
echo -e "${BOLD}${BLUE}    🚀 Velure - E-commerce Microservices${NC}"
echo -e "${BOLD}${BLUE}================================================${NC}"
echo ""

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker não está rodando!${NC}"
    echo -e "${YELLOW}Por favor, inicie o Docker Desktop e tente novamente.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker está rodando${NC}"
echo ""

# Verificar se .env existe
if [ ! -f "infrastructure/local/.env" ]; then
    echo -e "${YELLOW}⚠️  Arquivo .env não encontrado. Criando a partir do .env.example...${NC}"
    cp infrastructure/local/.env.example infrastructure/local/.env
    echo -e "${GREEN}✅ Arquivo .env criado${NC}"
    echo ""
fi

# Verificar entrada no /etc/hosts
if ! grep -q "velure.local" /etc/hosts; then
    echo -e "${YELLOW}⚠️  Entrada 'velure.local' não encontrada no /etc/hosts${NC}"
    echo -e "${YELLOW}Execute: echo '127.0.0.1 velure.local' | sudo tee -a /etc/hosts${NC}"
    echo ""
    read -p "Deseja adicionar agora? (s/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts > /dev/null
        echo -e "${GREEN}✅ Entrada adicionada ao /etc/hosts${NC}"
    fi
    echo ""
fi

# Detectar ambiente
KUBECTL_AVAILABLE=false
K6_AVAILABLE=false
KIND_AVAILABLE=false
KIND_CLUSTER_EXISTS=false

if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null 2>&1; then
    KUBECTL_AVAILABLE=true
    CLUSTER_NAME=$(kubectl config current-context 2>/dev/null || echo "unknown")
fi

if command -v k6 &> /dev/null; then
    K6_AVAILABLE=true
fi

if command -v kind &> /dev/null; then
    KIND_AVAILABLE=true
    if kind get clusters 2>/dev/null | grep -q "^velure$"; then
        KIND_CLUSTER_EXISTS=true
    fi
fi

# Menu de opções
echo -e "${BOLD}Escolha uma opção:${NC}"
echo ""
echo -e "${BOLD}${BLUE}━━━ Docker Local ━━━${NC}"
echo "  1) 🚀 Rodar TUDO (Aplicação + Grafana + Prometheus)"
echo "  2) 📦 Rodar apenas a Aplicação (sem monitoramento)"
echo "  3) 📊 Rodar apenas Monitoramento (Grafana + Prometheus)"
echo "  4) 🛑 Parar tudo (Docker)"
echo "  5) 🧹 Limpar tudo (remove containers e volumes)"
echo ""
echo -e "${BOLD}${BLUE}━━━ Kubernetes ━━━${NC}"
if [ "$KUBECTL_AVAILABLE" = true ]; then
    echo -e "  ${GREEN}✓${NC} Cluster: ${CLUSTER_NAME}"
    echo "  10) ☸️  Deploy no Kubernetes (completo)"
    echo "  11) 🗑️  Remover do Kubernetes"
    echo "  12) 📊 Ver status do Kubernetes"
    if [ "$KIND_CLUSTER_EXISTS" = true ]; then
        echo "  13) 🧹 Deletar cluster kind local"
    fi
else
    echo -e "  ${YELLOW}✗ kubectl não disponível ou cluster não conectado${NC}"
    echo ""
    if [ "$KIND_AVAILABLE" = true ]; then
        echo -e "  ${GREEN}✓${NC} kind está instalado"
        echo "  10) 🚀 Criar cluster Kubernetes local (kind) + Deploy completo"
        echo "  11) 📖 Ver instruções para configurar Kubernetes local"
    else
        echo -e "  ${YELLOW}ℹ️  Deseja rodar no Kubernetes localmente?${NC}"
        echo "  10) 📖 Ver como instalar kind (Kubernetes local)"
    fi
fi
echo ""
echo -e "${BOLD}${BLUE}━━━ Testes de Carga & HPA ━━━${NC}"
if [ "$K6_AVAILABLE" = true ] && [ "$KUBECTL_AVAILABLE" = true ]; then
    echo -e "  ${GREEN}✓${NC} k6 e kubectl disponíveis"
    echo "  20) 🧪 Teste de carga + Monitorar HPA (Kubernetes)"
    echo "  21) 📊 Apenas monitorar HPA em tempo real"
    echo "  22) 🎯 Teste integrado (todos serviços)"
    echo "  23) ⚙️  Instalar metrics-server (necessário para HPA)"
elif [ "$K6_AVAILABLE" = false ]; then
    echo -e "  ${YELLOW}✗ k6 não instalado${NC} (brew install k6)"
elif [ "$KUBECTL_AVAILABLE" = false ]; then
    echo -e "  ${YELLOW}✗ kubectl não disponível${NC}"
fi
echo ""
echo -e "${BOLD}${BLUE}━━━ Utilitários ━━━${NC}"
echo "  30) 📊 Abrir Grafana no navegador"
echo "  31) 🌐 Abrir Aplicação no navegador"
echo "  32) ❓ Ver status dos containers"
echo "  33) 📖 Ver ajuda/documentação"
echo ""
echo "  0) ❌ Sair"
echo ""

read -p "Digite sua escolha: " choice

case $choice in
    1)
        echo ""
        echo -e "${BLUE}🚀 Iniciando aplicação com monitoramento completo...${NC}"
        echo ""
        cd infrastructure/local
        docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml up -d
        echo ""
        echo -e "${GREEN}✅ Velure iniciado com sucesso!${NC}"
        echo ""
        echo -e "${BOLD}🌐 Acessos disponíveis:${NC}"
        echo ""
        echo -e "  Aplicação:    ${BLUE}https://velure.local${NC}"
        echo -e "  Grafana:      ${BLUE}http://localhost:3000${NC} (admin/admin)"
        echo -e "  Prometheus:   ${BLUE}http://localhost:9090${NC}"
        echo -e "  RabbitMQ:     ${BLUE}http://localhost:15672${NC} (admin/admin_password)"
        echo -e "  cAdvisor:     ${BLUE}http://localhost:8080${NC}"
        echo ""
        echo -e "${YELLOW}📊 Dashboard principal: http://localhost:3000/d/velure-overview${NC}"
        echo ""
        echo -e "${GREEN}Aguarde ~30 segundos para todos os serviços iniciarem completamente.${NC}"
        echo ""
        read -p "Pressione ENTER para abrir o Grafana no navegador..."
        open "http://localhost:3000/d/velure-overview" 2>/dev/null || xdg-open "http://localhost:3000/d/velure-overview" 2>/dev/null || echo "Abra manualmente: http://localhost:3000/d/velure-overview"
        ;;
    2)
        echo ""
        echo -e "${BLUE}📦 Iniciando apenas a aplicação...${NC}"
        echo ""
        cd infrastructure/local
        docker-compose up -d
        echo ""
        echo -e "${GREEN}✅ Aplicação iniciada!${NC}"
        echo -e "  Acesse: ${BLUE}https://velure.local${NC}"
        echo ""
        ;;
    3)
        echo ""
        echo -e "${BLUE}📊 Iniciando apenas monitoramento...${NC}"
        echo ""
        cd infrastructure/local
        docker-compose -f docker-compose.monitoring.yaml up -d
        echo ""
        echo -e "${GREEN}✅ Monitoramento iniciado!${NC}"
        echo -e "  Grafana: ${BLUE}http://localhost:3000${NC} (admin/admin)"
        echo -e "  Prometheus: ${BLUE}http://localhost:9090${NC}"
        echo ""
        ;;
    4)
        echo ""
        echo -e "${YELLOW}🛑 Parando todos os containers...${NC}"
        echo ""
        cd infrastructure/local
        docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml down
        echo ""
        echo -e "${GREEN}✅ Todos os containers foram parados${NC}"
        echo ""
        ;;
    5)
        echo ""
        echo -e "${RED}⚠️  ATENÇÃO: Isso irá remover todos os containers e volumes!${NC}"
        read -p "Tem certeza? (s/n) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Ss]$ ]]; then
            echo -e "${YELLOW}🧹 Limpando tudo...${NC}"
            cd infrastructure/local
            docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml down -v
            echo -e "${GREEN}✅ Limpeza concluída${NC}"
        else
            echo -e "${BLUE}Operação cancelada${NC}"
        fi
        echo ""
        ;;
    10)
        echo ""
        # Quando kubectl disponível: Deploy normal
        if [ "$KUBECTL_AVAILABLE" = true ]; then
            echo -e "${BLUE}☸️  Deploy completo no Kubernetes...${NC}"
            echo ""

            echo -e "${YELLOW}📦 1/4 - Criando namespaces...${NC}"
            kubectl create namespace datastores 2>/dev/null || echo "Namespace datastores já existe"
            kubectl create namespace velure 2>/dev/null || echo "Namespace velure já existe"

            echo -e "${YELLOW}📦 2/4 - Deploy datastores (MongoDB, Redis, RabbitMQ)...${NC}"
            helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
            helm repo update
            echo -e "${YELLOW}Baixando dependências do chart...${NC}"
            cd infrastructure/kubernetes/charts/velure-datastores
            helm dependency build .
            cd - > /dev/null
            helm upgrade --install velure-datastores infrastructure/kubernetes/charts/velure-datastores -n datastores

            echo -e "${YELLOW}⏳ Aguardando datastores ficarem prontos (30s)...${NC}"
            sleep 30

            echo -e "${YELLOW}📦 3/4 - Deploy serviços...${NC}"
            helm upgrade --install velure-auth infrastructure/kubernetes/charts/velure-auth -n velure
            helm upgrade --install velure-product infrastructure/kubernetes/charts/velure-product -n velure
            helm upgrade --install velure-publish-order infrastructure/kubernetes/charts/velure-publish-order -n velure
            helm upgrade --install velure-process-order infrastructure/kubernetes/charts/velure-process-order -n velure
            helm upgrade --install velure-ui infrastructure/kubernetes/charts/velure-ui -n velure

            echo -e "${YELLOW}📦 4/4 - Verificando status...${NC}"
            kubectl get pods -n velure
            echo ""
            echo -e "${GREEN}✅ Deploy concluído!${NC}"
            echo ""
            echo -e "${YELLOW}Para acessar, use port-forward:${NC}"
            echo "  kubectl port-forward -n velure svc/velure-ui 8080:80"
            echo ""

        # Quando kind disponível mas sem cluster: Criar cluster kind + Deploy
        elif [ "$KIND_AVAILABLE" = true ]; then
            echo -e "${BLUE}🚀 Criando cluster Kubernetes local (kind) + Deploy completo${NC}"
            echo ""
            ./scripts/k8s/setup-kind-cluster.sh

        # Quando nem kubectl nem kind: Mostrar instruções
        else
            echo -e "${BLUE}📖 Como instalar kind (Kubernetes local)${NC}"
            echo ""
            echo -e "${BOLD}kind${NC} permite rodar Kubernetes localmente usando Docker."
            echo ""
            echo -e "${YELLOW}Instalação (macOS):${NC}"
            echo "  brew install kind"
            echo "  brew install kubectl"
            echo "  brew install helm"
            echo ""
            echo -e "${YELLOW}Instalação (Linux):${NC}"
            echo "  # kind"
            echo "  curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64"
            echo "  chmod +x ./kind"
            echo "  sudo mv ./kind /usr/local/bin/kind"
            echo ""
            echo "  # kubectl"
            echo "  curl -LO https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
            echo "  chmod +x kubectl"
            echo "  sudo mv kubectl /usr/local/bin/"
            echo ""
            echo "  # helm"
            echo "  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
            echo ""
            echo -e "${YELLOW}Após instalar, execute novamente este script!${NC}"
            echo ""
            echo -e "${BLUE}Mais informações:${NC}"
            echo "  https://kind.sigs.k8s.io/docs/user/quick-start/"
            echo ""
        fi
        ;;
    11)
        echo ""
        # Quando kubectl disponível: Remover do K8s
        if [ "$KUBECTL_AVAILABLE" = true ]; then
            echo -e "${RED}⚠️  ATENÇÃO: Isso irá remover todos os recursos do Kubernetes!${NC}"
            read -p "Tem certeza? (s/n) " -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Ss]$ ]]; then
                echo -e "${YELLOW}🗑️  Removendo serviços...${NC}"
                helm uninstall velure-ui -n velure 2>/dev/null || true
                helm uninstall velure-process-order -n velure 2>/dev/null || true
                helm uninstall velure-publish-order -n velure 2>/dev/null || true
                helm uninstall velure-product -n velure 2>/dev/null || true
                helm uninstall velure-auth -n velure 2>/dev/null || true

                echo -e "${YELLOW}🗑️  Removendo datastores...${NC}"
                helm uninstall velure-datastores -n datastores 2>/dev/null || true

                echo -e "${YELLOW}🗑️  Removendo namespaces...${NC}"
                kubectl delete namespace velure 2>/dev/null || true
                kubectl delete namespace datastores 2>/dev/null || true

                echo -e "${GREEN}✅ Remoção concluída${NC}"
            else
                echo -e "${BLUE}Operação cancelada${NC}"
            fi

        # Quando kind disponível mas sem cluster: Mostrar instruções
        else
            echo -e "${BLUE}📖 Instruções para Kubernetes Local (kind)${NC}"
            echo ""
            echo -e "${BOLD}Pré-requisitos:${NC}"
            echo "  • Docker Desktop rodando"
            echo "  • kind instalado (brew install kind)"
            echo "  • kubectl instalado (brew install kubectl)"
            echo "  • helm instalado (brew install helm)"
            echo ""
            echo -e "${BOLD}Como usar:${NC}"
            echo "  1. Execute a opção 10 para criar cluster + deploy automático"
            echo "  2. Ou siga os passos manuais:"
            echo ""
            echo "     # Criar cluster"
            echo "     kind create cluster --config=infrastructure/kubernetes/kind-config.yaml"
            echo ""
            echo "     # Deploy completo"
            echo "     ./scripts/k8s/setup-kind-cluster.sh"
            echo ""
            echo -e "${BOLD}Documentação:${NC}"
            echo "  • https://kind.sigs.k8s.io/docs/user/quick-start/"
            echo ""
        fi
        echo ""
        ;;
    12)
        echo ""
        echo -e "${BLUE}📊 Status do Kubernetes${NC}"
        echo ""
        echo -e "${BOLD}Cluster:${NC} $(kubectl config current-context)"
        echo ""
        echo -e "${BOLD}━━━ Datastores (namespace: datastores) ━━━${NC}"
        kubectl get pods -n datastores 2>/dev/null || echo "Namespace não existe"
        echo ""
        echo -e "${BOLD}━━━ Serviços Velure (namespace: velure) ━━━${NC}"
        kubectl get pods -n velure 2>/dev/null || echo "Namespace não existe"
        echo ""
        echo -e "${BOLD}━━━ HPAs ━━━${NC}"
        kubectl get hpa -n velure 2>/dev/null || echo "Nenhum HPA encontrado"
        echo ""
        ;;
    13)
        echo ""
        echo -e "${BLUE}🧹 Deletar cluster kind local${NC}"
        echo ""
        if [ "$KIND_CLUSTER_EXISTS" = false ]; then
            echo -e "${YELLOW}⚠️  Nenhum cluster kind 'velure' encontrado${NC}"
            echo ""
            if kind get clusters 2>/dev/null | grep -q .; then
                echo -e "${YELLOW}Clusters kind disponíveis:${NC}"
                kind get clusters
            else
                echo -e "${YELLOW}Nenhum cluster kind encontrado${NC}"
            fi
            echo ""
        else
            echo -e "${RED}⚠️  ATENÇÃO: Isso irá deletar completamente o cluster kind 'velure'!${NC}"
            echo -e "${YELLOW}Todos os dados e configurações serão perdidos.${NC}"
            echo ""
            read -p "Tem certeza? (s/n) " -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Ss]$ ]]; then
                echo -e "${YELLOW}🗑️  Deletando cluster kind 'velure'...${NC}"
                kind delete cluster --name velure
                echo ""
                echo -e "${GREEN}✅ Cluster deletado com sucesso!${NC}"
                echo ""
                echo -e "${YELLOW}Para criar novamente, execute a opção 10${NC}"
            else
                echo -e "${BLUE}Operação cancelada${NC}"
            fi
        fi
        echo ""
        ;;
    20)
        echo ""
        echo -e "${BLUE}🧪 Teste de carga + Monitoramento HPA${NC}"
        echo ""
        if [ "$K6_AVAILABLE" = false ]; then
            echo -e "${RED}❌ k6 não instalado${NC}"
            echo -e "${YELLOW}Instale com: brew install k6${NC}"
            exit 1
        fi
        if [ "$KUBECTL_AVAILABLE" = false ]; then
            echo -e "${RED}❌ kubectl não disponível${NC}"
            exit 1
        fi

        echo -e "${YELLOW}Executando teste integrado...${NC}"
        echo -e "${YELLOW}Abra outro terminal e execute: ./tests/load/monitor-scaling.sh${NC}"
        echo ""
        read -p "Pressione ENTER para iniciar o teste..."

        cd tests/load
        ./run-k8s-local.sh integrated
        ;;
    21)
        echo ""
        echo -e "${BLUE}📊 Monitorando HPA em tempo real...${NC}"
        echo ""
        if [ "$KUBECTL_AVAILABLE" = false ]; then
            echo -e "${RED}❌ kubectl não disponível${NC}"
            exit 1
        fi

        cd tests/load
        ./monitor-scaling.sh
        ;;
    22)
        echo ""
        echo -e "${BLUE}🎯 Teste integrado (todos serviços)${NC}"
        echo ""
        if [ "$K6_AVAILABLE" = false ]; then
            echo -e "${RED}❌ k6 não instalado${NC}"
            echo -e "${YELLOW}Instale com: brew install k6${NC}"
            exit 1
        fi

        echo -e "${YELLOW}Escolha o ambiente:${NC}"
        echo "  1) Kubernetes local"
        echo "  2) Docker local (https://velure.local)"
        read -p "Digite sua escolha: " env_choice

        case $env_choice in
            1)
                cd tests/load
                ./run-k8s-local.sh integrated
                ;;
            2)
                if [ -f tests/load/.env.local ]; then
                    source tests/load/.env.local
                else
                    echo -e "${YELLOW}Usando configuração padrão...${NC}"
                    export BASE_URL=https://velure.local
                fi
                cd tests/load
                k6 run -e BASE_URL=$BASE_URL integrated-load-test.js
                ;;
            *)
                echo -e "${RED}Opção inválida${NC}"
                exit 1
                ;;
        esac
        ;;
    23)
        echo ""
        echo -e "${BLUE}⚙️  Instalando metrics-server...${NC}"
        echo ""
        if [ "$KUBECTL_AVAILABLE" = false ]; then
            echo -e "${RED}❌ kubectl não disponível${NC}"
            exit 1
        fi

        echo -e "${YELLOW}Aplicando metrics-server...${NC}"
        kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

        echo ""
        echo -e "${YELLOW}⏳ Aguardando metrics-server ficar pronto...${NC}"
        kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=60s

        echo ""
        echo -e "${GREEN}✅ Metrics-server instalado!${NC}"
        echo -e "${YELLOW}Testando: kubectl top nodes${NC}"
        sleep 5
        kubectl top nodes
        echo ""
        ;;
    30)
        echo ""
        echo -e "${BLUE}📊 Abrindo Grafana...${NC}"
        open "http://localhost:3000/d/velure-overview" 2>/dev/null || xdg-open "http://localhost:3000/d/velure-overview" 2>/dev/null || echo "Abra manualmente: http://localhost:3000/d/velure-overview"
        echo ""
        ;;
    31)
        echo ""
        echo -e "${BLUE}🌐 Abrindo aplicação...${NC}"
        open "https://velure.local" 2>/dev/null || xdg-open "https://velure.local" 2>/dev/null || echo "Abra manualmente: https://velure.local"
        echo ""
        ;;
    32)
        echo ""
        echo -e "${BLUE}📊 Status dos containers:${NC}"
        echo ""
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        ;;
    33)
        echo ""
        echo -e "${BLUE}📖 Documentação disponível:${NC}"
        echo ""
        echo -e "${BOLD}Arquivos principais:${NC}"
        echo "  - README.md - Documentação principal do projeto"
        echo "  - CLAUDE.md - Instruções para desenvolvimento"
        echo "  - docs/DEPLOY_GUIDE.md - Deploy AWS/EKS"
        echo "  - docs/LOAD_TESTING.md - Testes de carga e HPA"
        echo "  - docs/MONITORING.md - Monitoramento Kubernetes"
        echo "  - docs/TROUBLESHOOTING.md - Solução de problemas"
        echo ""
        echo -e "${BOLD}Comandos úteis:${NC}"
        echo "  - make help - Ver todos comandos do Makefile"
        echo "  - docker logs -f <container> - Logs em tempo real"
        echo "  - kubectl get pods -n velure - Ver pods no K8s"
        echo ""
        ;;
    0)
        echo ""
        echo -e "${BLUE}👋 Até logo!${NC}"
        echo ""
        exit 0
        ;;
    *)
        echo ""
        echo -e "${RED}❌ Opção inválida${NC}"
        echo ""
        exit 1
        ;;
esac

echo -e "${GREEN}Pronto! ✨${NC}"
echo ""

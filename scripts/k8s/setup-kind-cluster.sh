#!/bin/bash

# Velure - Setup Local Kubernetes Cluster with kind
# Este script cria um cluster kind completo com toda a aplicação Velure

set -e

BOLD="\033[1m"
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo ""
echo -e "${BOLD}${BLUE}================================================${NC}"
echo -e "${BOLD}${BLUE}    🚀 Velure - Setup Kubernetes Local (kind)${NC}"
echo -e "${BOLD}${BLUE}================================================${NC}"
echo ""

# Verificar se kind está instalado
if ! command -v kind &> /dev/null; then
    echo -e "${RED}❌ kind não está instalado${NC}"
    echo ""
    echo -e "${YELLOW}Instale o kind com:${NC}"
    echo "  brew install kind"
    echo ""
    echo "Ou visite: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Verificar se kubectl está instalado
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl não está instalado${NC}"
    echo ""
    echo -e "${YELLOW}Instale o kubectl com:${NC}"
    echo "  brew install kubectl"
    exit 1
fi

# Verificar se helm está instalado
if ! command -v helm &> /dev/null; then
    echo -e "${RED}❌ helm não está instalado${NC}"
    echo ""
    echo -e "${YELLOW}Instale o helm com:${NC}"
    echo "  brew install helm"
    exit 1
fi

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker não está rodando!${NC}"
    echo -e "${YELLOW}Por favor, inicie o Docker Desktop e tente novamente.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Pré-requisitos verificados (kind, kubectl, helm, docker)${NC}"
echo ""

# Verificar se já existe cluster velure
if kind get clusters 2>/dev/null | grep -q "^velure$"; then
    echo -e "${YELLOW}⚠️  Cluster 'velure' já existe${NC}"
    read -p "Deseja deletar e recriar? (s/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        echo -e "${YELLOW}🗑️  Deletando cluster existente...${NC}"
        kind delete cluster --name velure
    else
        echo -e "${BLUE}Usando cluster existente${NC}"
        kubectl cluster-info --context kind-velure
        echo ""
        read -p "Pressione ENTER para continuar com deploy..."
    fi
fi

# Criar cluster kind se não existir
if ! kind get clusters 2>/dev/null | grep -q "^velure$"; then
    echo -e "${BLUE}📦 1/8 - Criando cluster kind 'velure'...${NC}"
    kind create cluster --config="$REPO_ROOT/infrastructure/kubernetes/kind-config.yaml"
    echo -e "${GREEN}✅ Cluster criado${NC}"
    echo ""
fi

# Configurar kubectl context
kubectl cluster-info --context kind-velure
echo ""

# Instalar NGINX Ingress Controller
echo -e "${BLUE}📦 2/8 - Instalando NGINX Ingress Controller...${NC}"
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
helm repo update

# Verificar se já está instalado
if helm list -n ingress-nginx 2>/dev/null | grep -q "ingress-nginx"; then
    echo -e "${YELLOW}Ingress nginx já instalado, fazendo upgrade...${NC}"
fi

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=NodePort \
  --set controller.hostPort.enabled=true \
  --set controller.hostPort.ports.http=80 \
  --set controller.hostPort.ports.https=443 \
  --wait \
  --timeout=3m

echo -e "${GREEN}✅ NGINX Ingress instalado${NC}"
echo ""

# Criar namespaces
echo -e "${BLUE}📦 3/8 - Criando namespaces...${NC}"
kubectl create namespace datastores 2>/dev/null || echo "Namespace datastores já existe"
kubectl create namespace velure 2>/dev/null || echo "Namespace velure já existe"
echo -e "${GREEN}✅ Namespaces criados${NC}"
echo ""

# Deploy datastores
echo -e "${BLUE}📦 4/8 - Deploy datastores (MongoDB, Redis, RabbitMQ)...${NC}"
helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
helm repo update

DATASTORES_CHART="$REPO_ROOT/infrastructure/kubernetes/charts/velure-datastores"

# Deploy usando --dependency-update para resolver dependências automaticamente
helm upgrade --install velure-datastores "$DATASTORES_CHART" \
  -n datastores \
  --dependency-update \
  --wait \
  --timeout=5m

echo -e "${GREEN}✅ Datastores deployados${NC}"
echo ""

# Verificar se imagens Docker existem
echo -e "${BLUE}📦 5/8 - Verificando imagens Docker...${NC}"

IMAGES=(
    "velure-auth-service:latest"
    "velure-product-service:latest"
    "velure-publish-order-service:latest"
    "velure-process-order-service:latest"
    "velure-ui-service:latest"
)

MISSING_IMAGES=()
for img in "${IMAGES[@]}"; do
    if ! docker image inspect "$img" > /dev/null 2>&1; then
        MISSING_IMAGES+=("$img")
    fi
done

if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    echo -e "${YELLOW}⚠️  As seguintes imagens não foram encontradas:${NC}"
    for img in "${MISSING_IMAGES[@]}"; do
        echo "  - $img"
    done
    echo ""
    echo -e "${YELLOW}Deseja fazer build agora? (Isso pode demorar alguns minutos)${NC}"
    read -p "(s/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        echo -e "${BLUE}Building imagens...${NC}"
        cd "$REPO_ROOT"
        make docker-build
        echo -e "${GREEN}✅ Build concluído${NC}"
    else
        echo -e "${RED}❌ Imagens necessárias não disponíveis. Abortando.${NC}"
        exit 1
    fi
fi

# Carregar imagens no kind
echo -e "${BLUE}📦 6/8 - Carregando imagens no cluster kind...${NC}"
for img in "${IMAGES[@]}"; do
    echo "  Carregando $img..."
    kind load docker-image "$img" --name velure
done
echo -e "${GREEN}✅ Imagens carregadas${NC}"
echo ""

# Deploy serviços
echo -e "${BLUE}📦 7/8 - Deploy dos microserviços...${NC}"

# Auth Service
echo "  → velure-auth"
helm upgrade --install velure-auth \
  "$REPO_ROOT/infrastructure/kubernetes/charts/velure-auth" \
  -n velure \
  --set image.pullPolicy=Never \
  --set ingress.className=nginx \
  --wait --timeout=2m

# Product Service
echo "  → velure-product"
helm upgrade --install velure-product \
  "$REPO_ROOT/infrastructure/kubernetes/charts/velure-product" \
  -n velure \
  --set image.pullPolicy=Never \
  --set ingress.className=nginx \
  --wait --timeout=2m

# Publish Order Service
echo "  → velure-publish-order"
helm upgrade --install velure-publish-order \
  "$REPO_ROOT/infrastructure/kubernetes/charts/velure-publish-order" \
  -n velure \
  --set image.pullPolicy=Never \
  --wait --timeout=2m

# Process Order Service
echo "  → velure-process-order"
helm upgrade --install velure-process-order \
  "$REPO_ROOT/infrastructure/kubernetes/charts/velure-process-order" \
  -n velure \
  --set image.pullPolicy=Never \
  --wait --timeout=2m

# UI Service
echo "  → velure-ui"
helm upgrade --install velure-ui \
  "$REPO_ROOT/infrastructure/kubernetes/charts/velure-ui" \
  -n velure \
  --set image.pullPolicy=Never \
  --set ingress.className=nginx \
  --wait --timeout=2m

echo -e "${GREEN}✅ Microserviços deployados${NC}"
echo ""

# Configurar /etc/hosts
echo -e "${BLUE}📦 8/8 - Configurando /etc/hosts...${NC}"

HOSTS_ENTRIES=(
    "velure.local"
    "auth.velure.local"
    "product.velure.local"
)

MISSING_HOSTS=()
for host in "${HOSTS_ENTRIES[@]}"; do
    if ! grep -q "$host" /etc/hosts; then
        MISSING_HOSTS+=("$host")
    fi
done

if [ ${#MISSING_HOSTS[@]} -gt 0 ]; then
    echo -e "${YELLOW}As seguintes entradas precisam ser adicionadas ao /etc/hosts:${NC}"
    for host in "${MISSING_HOSTS[@]}"; do
        echo "  127.0.0.1 $host"
    done
    echo ""
    read -p "Deseja adicionar agora? (requer sudo) (s/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        for host in "${MISSING_HOSTS[@]}"; do
            echo "127.0.0.1 $host" | sudo tee -a /etc/hosts > /dev/null
        done
        echo -e "${GREEN}✅ Entradas adicionadas ao /etc/hosts${NC}"
    else
        echo -e "${YELLOW}⚠️  Você precisará adicionar manualmente:${NC}"
        for host in "${MISSING_HOSTS[@]}"; do
            echo "  echo '127.0.0.1 $host' | sudo tee -a /etc/hosts"
        done
    fi
else
    echo -e "${GREEN}✅ /etc/hosts já configurado${NC}"
fi
echo ""

# Verificar status
echo -e "${BLUE}📊 Status do deployment:${NC}"
echo ""
kubectl get pods -n datastores
echo ""
kubectl get pods -n velure
echo ""
kubectl get ingress -n velure
echo ""

echo -e "${BOLD}${GREEN}🎉 Setup concluído com sucesso!${NC}"
echo ""
echo -e "${BOLD}🌐 Acessos disponíveis:${NC}"
echo ""
echo -e "  Aplicação:    ${BLUE}http://velure.local${NC}"
echo -e "  Auth API:     ${BLUE}http://auth.velure.local${NC}"
echo -e "  Product API:  ${BLUE}http://product.velure.local${NC}"
echo ""
echo -e "${BOLD}📊 Comandos úteis:${NC}"
echo ""
echo "  # Ver pods"
echo "  kubectl get pods -n velure"
echo ""
echo "  # Ver logs de um serviço"
echo "  kubectl logs -f deployment/velure-auth -n velure"
echo ""
echo "  # Ver status dos ingress"
echo "  kubectl get ingress -n velure"
echo ""
echo "  # Deletar cluster"
echo "  kind delete cluster --name velure"
echo ""
echo -e "${GREEN}Pronto! ✨${NC}"
echo ""

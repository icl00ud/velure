#!/bin/bash

# ======================================================================
# Script: 02-install-datastores.sh
# Descrição: Instala datastores (MongoDB, Redis, RabbitMQ) no cluster
# ======================================================================

set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Verificações
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl não encontrado"
    exit 1
fi

if ! command -v helm &> /dev/null; then
    log_error "helm não encontrado"
    exit 1
fi

# Verificar conexão
log_info "Verificando conexão com cluster..."
if ! kubectl cluster-info &> /dev/null; then
    log_error "Não conectado ao cluster"
    exit 1
fi

CLUSTER_NAME=$(kubectl config current-context | awk -F'/' '{print $2}')
log_info "Conectado ao cluster: $CLUSTER_NAME"

# ======================================================================
# Configuração
# ======================================================================

NAMESPACE="datastores"
CHART_PATH="../../infrastructure/kubernetes/charts/velure-datastores"

# Verificar se chart existe
if [ ! -d "$CHART_PATH" ]; then
    log_error "Chart não encontrado em: $CHART_PATH"
    log_info "Execute este script do diretório: velure/scripts/deploy/"
    exit 1
fi

# ======================================================================
# Criar Namespace
# ======================================================================

log_info "=== Preparando namespace $NAMESPACE ==="

if kubectl get namespace "$NAMESPACE" &> /dev/null; then
    log_warn "Namespace $NAMESPACE já existe"
else
    log_info "Criando namespace $NAMESPACE..."
    kubectl create namespace "$NAMESPACE"
    log_info "✓ Namespace criado"
fi

# Label namespace para Prometheus scraping
kubectl label namespace "$NAMESPACE" monitoring=enabled --overwrite

# ======================================================================
# Adicionar repositórios Helm
# ======================================================================

log_info "=== Configurando repositórios Helm ==="

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

log_info "✓ Repositórios atualizados"

# ======================================================================
# Atualizar dependências do chart
# ======================================================================

log_info "=== Atualizando dependências do chart ==="

cd "$CHART_PATH"
helm dependency update
cd - > /dev/null

log_info "✓ Dependências atualizadas"

# ======================================================================
# Instalar/Atualizar Datastores
# ======================================================================

log_info "=== Instalando Velure Datastores ==="

RELEASE_NAME="velure-datastores"

if helm list -n "$NAMESPACE" | grep -q "$RELEASE_NAME"; then
    log_warn "Release $RELEASE_NAME já existe, atualizando..."

    helm upgrade "$RELEASE_NAME" "$CHART_PATH" \
        --namespace "$NAMESPACE" \
        --timeout 10m \
        --wait

    log_info "✓ Release atualizada"
else
    log_info "Instalando release $RELEASE_NAME..."

    helm install "$RELEASE_NAME" "$CHART_PATH" \
        --namespace "$NAMESPACE" \
        --create-namespace \
        --timeout 10m \
        --wait

    log_info "✓ Release instalada"
fi

# ======================================================================
# Verificar instalação
# ======================================================================

log_info "=== Verificando instalação ==="

echo ""
log_info "Pods:"
kubectl get pods -n "$NAMESPACE"

echo ""
log_info "Services:"
kubectl get svc -n "$NAMESPACE"

echo ""
log_info "PersistentVolumeClaims:"
kubectl get pvc -n "$NAMESPACE"

# Aguardar todos os pods ficarem prontos
log_info "Aguardando todos os pods ficarem prontos..."

kubectl wait --for=condition=ready pod \
    -l app.kubernetes.io/instance=velure-datastores \
    -n "$NAMESPACE" \
    --timeout=600s || log_warn "Alguns pods podem ainda estar inicializando"

# ======================================================================
# Informações de Conexão
# ======================================================================

echo ""
log_info "=== Informações de Conexão ==="

echo ""
log_info "MongoDB:"
echo "  Internal URL: mongodb://productuser:product_password@velure-datastores-mongodb:27017/productdb"
echo "  Port-forward:  kubectl port-forward -n $NAMESPACE svc/velure-datastores-mongodb 27017:27017"
echo "  Connect CLI:   kubectl exec -it -n $NAMESPACE velure-datastores-mongodb-0 -- mongosh"

echo ""
log_info "Redis:"
echo "  Internal URL: redis://:redis_password@velure-datastores-redis-master:6379"
echo "  Port-forward:  kubectl port-forward -n $NAMESPACE svc/velure-datastores-redis-master 6379:6379"
echo "  Connect CLI:   kubectl exec -it -n $NAMESPACE velure-datastores-redis-master-0 -- redis-cli -a redis_password"

echo ""
log_info "RabbitMQ:"
echo "  AMQP URL:      amqp://publisher-order:publisher_password@velure-datastores-rabbitmq:5672/"
echo "  Management UI: kubectl port-forward -n $NAMESPACE svc/velure-datastores-rabbitmq 15672:15672"
echo "  Admin User:    admin / admin_password"
echo "  Access UI:     http://localhost:15672"

# ======================================================================
# Teste de Conectividade
# ======================================================================

echo ""
log_info "=== Testando Conectividade ==="

# Testar MongoDB
log_info "Testando MongoDB..."
if kubectl exec -n "$NAMESPACE" velure-datastores-mongodb-0 -- \
    mongosh --quiet --eval "db.adminCommand('ping')" &> /dev/null; then
    log_info "✓ MongoDB respondendo"
else
    log_warn "MongoDB pode não estar pronto ainda"
fi

# Testar Redis
log_info "Testando Redis..."
if kubectl exec -n "$NAMESPACE" velure-datastores-redis-master-0 -- \
    redis-cli -a redis_password ping 2>/dev/null | grep -q PONG; then
    log_info "✓ Redis respondendo"
else
    log_warn "Redis pode não estar pronto ainda"
fi

# Testar RabbitMQ
log_info "Testando RabbitMQ..."
if kubectl exec -n "$NAMESPACE" velure-datastores-rabbitmq-0 -- \
    rabbitmqctl status &> /dev/null; then
    log_info "✓ RabbitMQ respondendo"
else
    log_warn "RabbitMQ pode não estar pronto ainda"
fi

# ======================================================================
# Conclusão
# ======================================================================

echo ""
log_info "✓ Datastores instalados com sucesso!"
echo ""
log_info "Próximos passos:"
log_info "  1. Execute: ./03-install-monitoring.sh"
log_info "  2. Execute: ./04-deploy-services.sh"
echo ""
log_info "Para verificar status:"
log_info "  kubectl get all -n $NAMESPACE"
log_info "  helm status $RELEASE_NAME -n $NAMESPACE"

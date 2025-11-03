#!/bin/bash

# ======================================================================
# Script: 04-deploy-services.sh
# Descrição: Deploy dos microserviços Velure no Kubernetes
# ======================================================================

set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_blue() { echo -e "${BLUE}[INFO]${NC} $1"; }

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

NAMESPACE="default"
CHARTS_DIR="../../infrastructure/kubernetes/charts"

# Verificar diretório de charts
if [ ! -d "$CHARTS_DIR" ]; then
    log_error "Diretório de charts não encontrado: $CHARTS_DIR"
    exit 1
fi

# Serviços para deploy
SERVICES=("velure-auth" "velure-product" "velure-publish-order" "velure-process-order" "velure-ui")

# ======================================================================
# Criar Secrets
# ======================================================================

log_info "=== Criando Secrets ==="

# Secret para JWT
if kubectl get secret jwt-secret -n "$NAMESPACE" &> /dev/null; then
    log_warn "Secret jwt-secret já existe"
else
    log_info "Criando secret jwt-secret..."
    kubectl create secret generic jwt-secret \
        --from-literal=jwt-secret="your-super-secret-jwt-key-change-in-production" \
        --from-literal=jwt-refresh-secret="your-refresh-secret-key" \
        -n "$NAMESPACE"
    log_info "✓ Secret jwt-secret criado"
fi

# Secret para database connections
if kubectl get secret database-secrets -n "$NAMESPACE" &> /dev/null; then
    log_warn "Secret database-secrets já existe"
else
    log_info "Criando secret database-secrets..."
    kubectl create secret generic database-secrets \
        --from-literal=postgres-auth-url="postgresql://postgres:postgres_password@velure-rds-auth.xxxxx.us-east-1.rds.amazonaws.com:5432/authdb" \
        --from-literal=postgres-orders-url="postgresql://postgres:postgres_password@velure-rds-orders.xxxxx.us-east-1.rds.amazonaws.com:5432/ordersdb" \
        --from-literal=mongodb-url="mongodb://productuser:product_password@velure-datastores-mongodb:27017/productdb" \
        --from-literal=redis-url="redis://:redis_password@velure-datastores-redis-master:6379" \
        --from-literal=rabbitmq-url="amqp://publisher-order:publisher_password@velure-datastores-rabbitmq:5672/" \
        -n "$NAMESPACE"
    log_info "✓ Secret database-secrets criado"
fi

log_warn "IMPORTANTE: Edite os secrets com os valores corretos antes de continuar!"
log_warn "Execute: kubectl edit secret database-secrets -n $NAMESPACE"
echo ""
read -p "Pressione ENTER para continuar após editar os secrets..."

# ======================================================================
# Deploy dos Serviços
# ======================================================================

log_info "=== Deployando Serviços ==="

for SERVICE in "${SERVICES[@]}"; do
    CHART_PATH="$CHARTS_DIR/$SERVICE"

    if [ ! -d "$CHART_PATH" ]; then
        log_warn "Chart não encontrado: $CHART_PATH (pulando)"
        continue
    fi

    echo ""
    log_blue "Deployando $SERVICE..."

    if helm list -n "$NAMESPACE" | grep -q "^$SERVICE"; then
        log_info "Atualizando release existente..."
        helm upgrade "$SERVICE" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            --timeout 5m \
            --wait || log_error "Falha ao atualizar $SERVICE"
    else
        log_info "Instalando nova release..."
        helm install "$SERVICE" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            --create-namespace \
            --timeout 5m \
            --wait || log_error "Falha ao instalar $SERVICE"
    fi

    log_info "✓ $SERVICE deployado"
done

# ======================================================================
# Aplicar ServiceMonitors (se ainda não aplicados)
# ======================================================================

log_info "=== Aplicando ServiceMonitors ==="

SERVICEMONITORS_DIR="../../infrastructure/kubernetes/monitoring/servicemonitors"
if [ -d "$SERVICEMONITORS_DIR" ]; then
    kubectl apply -f "$SERVICEMONITORS_DIR" || log_warn "Alguns ServiceMonitors falharam"
    log_info "✓ ServiceMonitors aplicados"
fi

# ======================================================================
# Verificar Deployments
# ======================================================================

log_info "=== Verificando Deployments ==="

echo ""
log_info "Pods:"
kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance in (velure-auth,velure-product,velure-publish-order,velure-process-order,velure-ui)"

echo ""
log_info "Services:"
kubectl get svc -n "$NAMESPACE" -l "app.kubernetes.io/instance in (velure-auth,velure-product,velure-publish-order,velure-process-order,velure-ui)"

echo ""
log_info "Ingresses:"
kubectl get ingress -n "$NAMESPACE"

# Aguardar pods ficarem prontos
log_info "Aguardando todos os pods ficarem prontos..."
for SERVICE in "${SERVICES[@]}"; do
    kubectl wait --for=condition=ready pod \
        -l "app.kubernetes.io/name=$SERVICE" \
        -n "$NAMESPACE" \
        --timeout=300s || log_warn "$SERVICE pode estar com problemas"
done

# ======================================================================
# Obter URLs de Acesso
# ======================================================================

echo ""
log_info "=== URLs de Acesso ==="

# Obter LoadBalancer URL do Ingress (se existir)
INGRESS_URL=$(kubectl get ingress -n "$NAMESPACE" velure-ui -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)

if [ -z "$INGRESS_URL" ]; then
    log_warn "Ingress ainda sem External URL. Aguarde alguns minutos e execute:"
    echo "  kubectl get ingress -n $NAMESPACE velure-ui"
else
    echo ""
    log_blue "Aplicação acessível em:"
    echo "  http://$INGRESS_URL"
fi

# Endpoints individuais
echo ""
log_blue "Endpoints dos serviços:"
echo "  Auth Service:    /api/auth/*"
echo "  Product Service: /api/product/*"
echo "  Order Service:   /api/order/*"
echo "  UI:              /"

# ======================================================================
# Verificar Prometheus Targets
# ======================================================================

echo ""
log_info "=== Verificando Prometheus Targets ==="

PROMETHEUS_SVC=$(kubectl get svc -n monitoring -l "app.kubernetes.io/name=prometheus" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -n "$PROMETHEUS_SVC" ]; then
    log_info "Para verificar se as métricas estão sendo coletadas:"
    echo "  kubectl port-forward -n monitoring svc/$PROMETHEUS_SVC 9090:9090"
    echo "  Acesse: http://localhost:9090/targets"
    echo "  Procure por: velure-auth, velure-product, velure-publish-order, velure-process-order"
else
    log_warn "Prometheus não encontrado. Execute: ./03-install-monitoring.sh"
fi

# ======================================================================
# Health Checks
# ======================================================================

echo ""
log_info "=== Health Checks ==="

for SERVICE in "${SERVICES[@]}"; do
    SERVICE_NAME="${SERVICE/velure-/}"
    POD=$(kubectl get pod -n "$NAMESPACE" -l "app.kubernetes.io/name=$SERVICE" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

    if [ -n "$POD" ]; then
        log_info "Testando $SERVICE..."
        if kubectl exec -n "$NAMESPACE" "$POD" -- wget -q -O- http://localhost:8080/health 2>/dev/null | grep -q "ok"; then
            log_info "✓ $SERVICE está saudável"
        else
            log_warn "$SERVICE pode ter problemas de saúde"
        fi
    fi
done

# ======================================================================
# Conclusão
# ======================================================================

echo ""
log_info "✓ Deploy dos serviços concluído!"
echo ""
log_info "Próximos passos:"
log_info "  1. Aguarde o LoadBalancer obter External-IP"
log_info "  2. Acesse a aplicação no navegador"
log_info "  3. Verifique logs: kubectl logs -f -n $NAMESPACE <pod-name>"
log_info "  4. Verifique métricas no Prometheus e Grafana"
echo ""
log_info "Comandos úteis:"
log_info "  # Ver todos os recursos"
log_info "  kubectl get all -n $NAMESPACE"
echo ""
log_info "  # Logs de um serviço"
log_info "  kubectl logs -f -n $NAMESPACE -l app.kubernetes.io/name=velure-auth"
echo ""
log_info "  # Descrever pod com problemas"
log_info "  kubectl describe pod -n $NAMESPACE <pod-name>"
echo ""
log_info "  # Restart de um deployment"
log_info "  kubectl rollout restart deployment/velure-auth -n $NAMESPACE"

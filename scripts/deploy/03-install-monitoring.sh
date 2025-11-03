#!/bin/bash

# ======================================================================
# Script: 03-install-monitoring.sh
# Descrição: Instala stack de monitoramento (Prometheus + Grafana)
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

NAMESPACE="monitoring"
RELEASE_NAME="kube-prometheus-stack"
VALUES_FILE="../../infrastructure/kubernetes/monitoring/kube-prometheus-stack-values.yaml"
SERVICEMONITORS_DIR="../../infrastructure/kubernetes/monitoring/servicemonitors"

# Verificar arquivos
if [ ! -f "$VALUES_FILE" ]; then
    log_error "Values file não encontrado: $VALUES_FILE"
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

# ======================================================================
# Adicionar Repositório Helm
# ======================================================================

log_info "=== Configurando repositório Helm ==="

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

log_info "✓ Repositório atualizado"

# ======================================================================
# Instalar CRDs (se não existirem)
# ======================================================================

log_info "=== Verificando CRDs do Prometheus Operator ==="

if kubectl get crd prometheuses.monitoring.coreos.com &> /dev/null; then
    log_warn "CRDs já instalados"
else
    log_info "Instalando CRDs..."
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
    kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.70.0/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml
    log_info "✓ CRDs instalados"
fi

# ======================================================================
# Instalar/Atualizar kube-prometheus-stack
# ======================================================================

log_info "=== Instalando kube-prometheus-stack ==="

if helm list -n "$NAMESPACE" | grep -q "$RELEASE_NAME"; then
    log_warn "Release $RELEASE_NAME já existe, atualizando..."

    helm upgrade "$RELEASE_NAME" prometheus-community/kube-prometheus-stack \
        --namespace "$NAMESPACE" \
        --values "$VALUES_FILE" \
        --timeout 10m \
        --wait

    log_info "✓ Release atualizada"
else
    log_info "Instalando release $RELEASE_NAME..."

    helm install "$RELEASE_NAME" prometheus-community/kube-prometheus-stack \
        --namespace "$NAMESPACE" \
        --create-namespace \
        --values "$VALUES_FILE" \
        --timeout 10m \
        --wait

    log_info "✓ Release instalada"
fi

# ======================================================================
# Aplicar ServiceMonitors
# ======================================================================

log_info "=== Aplicando ServiceMonitors ==="

if [ -d "$SERVICEMONITORS_DIR" ]; then
    kubectl apply -f "$SERVICEMONITORS_DIR" || log_warn "Alguns ServiceMonitors podem ter falhado"
    log_info "✓ ServiceMonitors aplicados"
else
    log_warn "Diretório de ServiceMonitors não encontrado: $SERVICEMONITORS_DIR"
fi

# ======================================================================
# Verificar Instalação
# ======================================================================

log_info "=== Verificando instalação ==="

echo ""
log_info "Pods no namespace $NAMESPACE:"
kubectl get pods -n "$NAMESPACE"

echo ""
log_info "Services:"
kubectl get svc -n "$NAMESPACE"

# Aguardar pods ficarem prontos
log_info "Aguardando pods ficarem prontos..."
kubectl wait --for=condition=ready pod \
    -l "release=$RELEASE_NAME" \
    -n "$NAMESPACE" \
    --timeout=600s || log_warn "Alguns pods podem estar inicializando"

# ======================================================================
# Informações de Acesso
# ======================================================================

echo ""
log_info "=== Informações de Acesso ==="

# Prometheus
PROMETHEUS_SVC=$(kubectl get svc -n "$NAMESPACE" -l "app.kubernetes.io/name=prometheus" -o jsonpath='{.items[0].metadata.name}')
echo ""
log_blue "Prometheus:"
echo "  Port-forward: kubectl port-forward -n $NAMESPACE svc/$PROMETHEUS_SVC 9090:9090"
echo "  Access:       http://localhost:9090"

# Grafana
GRAFANA_SVC=$(kubectl get svc -n "$NAMESPACE" -l "app.kubernetes.io/name=grafana" -o jsonpath='{.items[0].metadata.name}')
GRAFANA_TYPE=$(kubectl get svc -n "$NAMESPACE" "$GRAFANA_SVC" -o jsonpath='{.spec.type}')

echo ""
log_blue "Grafana:"
echo "  Type: $GRAFANA_TYPE"

if [ "$GRAFANA_TYPE" = "LoadBalancer" ]; then
    log_info "Aguardando LoadBalancer obter External-IP..."
    GRAFANA_URL=$(kubectl get svc -n "$NAMESPACE" "$GRAFANA_SVC" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

    if [ -z "$GRAFANA_URL" ]; then
        log_warn "LoadBalancer ainda sem External-IP. Execute para verificar:"
        echo "  kubectl get svc -n $NAMESPACE $GRAFANA_SVC -w"
    else
        echo "  URL:      http://$GRAFANA_URL"
    fi
else
    echo "  Port-forward: kubectl port-forward -n $NAMESPACE svc/$GRAFANA_SVC 3000:80"
    echo "  Access:       http://localhost:3000"
fi

echo "  Username: admin"
echo "  Password: admin"

# Alertmanager
ALERTMANAGER_SVC=$(kubectl get svc -n "$NAMESPACE" -l "app.kubernetes.io/name=alertmanager" -o jsonpath='{.items[0].metadata.name}')
echo ""
log_blue "Alertmanager:"
echo "  Port-forward: kubectl port-forward -n $NAMESPACE svc/$ALERTMANAGER_SVC 9093:9093"
echo "  Access:       http://localhost:9093"

# ======================================================================
# Verificar ServiceMonitors
# ======================================================================

echo ""
log_info "=== ServiceMonitors Ativos ==="
kubectl get servicemonitor -A

# ======================================================================
# Verificar PrometheusRules (Alertas)
# ======================================================================

echo ""
log_info "=== PrometheusRules (Alertas) ==="
kubectl get prometheusrule -n "$NAMESPACE"

# ======================================================================
# Teste de Conectividade
# ======================================================================

echo ""
log_info "=== Testando Conectividade ==="

# Testar Prometheus
log_info "Testando Prometheus..."
if kubectl exec -n "$NAMESPACE" "$PROMETHEUS_SVC-0" -- wget -q -O- http://localhost:9090/-/healthy | grep -q "Prometheus"; then
    log_info "✓ Prometheus respondendo"
else
    log_warn "Prometheus pode não estar pronto"
fi

# ======================================================================
# Conclusão
# ======================================================================

echo ""
log_info "✓ Stack de monitoramento instalado com sucesso!"
echo ""
log_info "Próximos passos:"
log_info "  1. Acesse o Grafana e verifique se o datasource Prometheus está configurado"
log_info "  2. Execute: ./04-deploy-services.sh"
log_info "  3. Após deploy dos serviços, verifique métricas em Prometheus"
echo ""
log_info "Comandos úteis:"
log_info "  # Ver targets do Prometheus"
log_info "  kubectl port-forward -n $NAMESPACE svc/$PROMETHEUS_SVC 9090:9090"
log_info "  # Acesse: http://localhost:9090/targets"
echo ""
log_info "  # Ver alertas ativos"
log_info "  kubectl port-forward -n $NAMESPACE svc/$ALERTMANAGER_SVC 9093:9093"
log_info "  # Acesse: http://localhost:9093"
echo ""
log_info "  # Status do stack"
log_info "  helm status $RELEASE_NAME -n $NAMESPACE"

#!/bin/bash

# ========================================================================
# Script: run-k8s-local.sh
# Description: Run k6 load tests against local Kubernetes cluster
# Usage: ./run-k8s-local.sh [test-name]
#        test-name: auth | product | order | ui | integrated | all
# ========================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# ========================================================================
# Configuration
# ========================================================================

TEST_NAME="${1:-integrated}"
NAMESPACE="${NAMESPACE:-default}"
WARMUP_DURATION="${WARMUP_DURATION:-30s}"
TEST_DURATION="${TEST_DURATION:-15s}"

# ========================================================================
# Validate Prerequisites
# ========================================================================

log_step "Validating prerequisites..."

if ! command -v k6 &> /dev/null; then
    log_error "k6 is not installed"
    log_info "Install: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    log_error "kubectl is not installed"
    exit 1
fi

# Check cluster connection
if ! kubectl cluster-info &> /dev/null; then
    log_error "Not connected to Kubernetes cluster"
    log_info "Ensure your cluster is running (minikube, kind, etc.)"
    exit 1
fi

CLUSTER_NAME=$(kubectl config current-context)
log_info "Connected to cluster: $CLUSTER_NAME"

# ========================================================================
# Check HPA Support
# ========================================================================

log_step "Checking HPA support..."

if ! kubectl get hpa -n $NAMESPACE &> /dev/null; then
    log_warn "HPA resources not found in namespace $NAMESPACE"
    log_info "Make sure services are deployed with HPA enabled"
fi

# Check metrics-server
if ! kubectl get deployment metrics-server -n kube-system &> /dev/null; then
    log_warn "metrics-server not found"
    log_info "HPA requires metrics-server. Install with:"
    log_info "  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml"
fi

# ========================================================================
# Detect Service URLs
# ========================================================================

log_step "Detecting service URLs..."

# Try to get Ingress URL
INGRESS_HOST=$(kubectl get ingress -n $NAMESPACE -o jsonpath='{.items[0].spec.rules[0].host}' 2>/dev/null || echo "")

if [ -n "$INGRESS_HOST" ]; then
    log_info "Found Ingress: $INGRESS_HOST"
    BASE_URL="http://$INGRESS_HOST"
    AUTH_URL="$BASE_URL/api/auth"
    PRODUCT_URL="$BASE_URL/api/product"
    ORDER_URL="$BASE_URL/api/order"
    UI_URL="$BASE_URL"
else
    log_warn "No Ingress found, using port-forward"

    # Port-forward services
    log_info "Setting up port-forwards..."

    # Kill existing port-forwards
    pkill -f "kubectl port-forward" 2>/dev/null || true
    sleep 2

    # Start port-forwards in background
    kubectl port-forward -n $NAMESPACE svc/velure-auth 3020:3020 &> /dev/null &
    kubectl port-forward -n $NAMESPACE svc/velure-product 3010:3010 &> /dev/null &
    kubectl port-forward -n $NAMESPACE svc/velure-publish-order 8080:8080 &> /dev/null &
    kubectl port-forward -n $NAMESPACE svc/velure-ui 8081:80 &> /dev/null &

    sleep 5

    AUTH_URL="http://localhost:3020"
    PRODUCT_URL="http://localhost:3010"
    ORDER_URL="http://localhost:8080"
    UI_URL="http://localhost:8081"
fi

log_info "Service URLs:"
log_info "  Auth:    $AUTH_URL"
log_info "  Product: $PRODUCT_URL"
log_info "  Order:   $ORDER_URL"
log_info "  UI:      $UI_URL"

# ========================================================================
# Health Check
# ========================================================================

log_step "Checking service health..."

check_health() {
    local url=$1
    local name=$2

    if curl -f -s -o /dev/null "$url/health" 2>/dev/null || curl -f -s -o /dev/null "$url" 2>/dev/null; then
        log_info "✓ $name is healthy"
        return 0
    else
        log_warn "✗ $name is not responding"
        return 1
    fi
}

# Wait for services to be ready
sleep 3

# Check each service
check_health "$AUTH_URL" "Auth Service" || log_warn "Auth service may not be ready"
check_health "$PRODUCT_URL" "Product Service" || log_warn "Product service may not be ready"
check_health "$ORDER_URL" "Order Service" || log_warn "Order service may not be ready"

# ========================================================================
# Show HPA Status Before Test
# ========================================================================

log_step "Current HPA status:"
kubectl get hpa -n $NAMESPACE 2>/dev/null || log_info "No HPA resources found"

# ========================================================================
# Run K6 Test
# ========================================================================

log_step "Running k6 load test: $TEST_NAME"

case $TEST_NAME in
    auth)
        TEST_FILE="auth-service-test.js"
        k6 run \
            -e AUTH_URL="$AUTH_URL" \
            -e WARMUP_DURATION="$WARMUP_DURATION" \
            -e TEST_DURATION="$TEST_DURATION" \
            "$TEST_FILE"
        ;;
    product)
        TEST_FILE="product-service-test.js"
        k6 run \
            -e PRODUCT_URL="$PRODUCT_URL" \
            -e WARMUP_DURATION="$WARMUP_DURATION" \
            -e TEST_DURATION="$TEST_DURATION" \
            "$TEST_FILE"
        ;;
    order)
        TEST_FILE="publish-order-service-test.js"
        k6 run \
            -e ORDER_URL="$ORDER_URL" \
            -e WARMUP_DURATION="$WARMUP_DURATION" \
            -e TEST_DURATION="$TEST_DURATION" \
            "$TEST_FILE"
        ;;
    ui)
        TEST_FILE="ui-service-test.js"
        k6 run \
            -e UI_URL="$UI_URL" \
            -e WARMUP_DURATION="$WARMUP_DURATION" \
            -e TEST_DURATION="$TEST_DURATION" \
            "$TEST_FILE"
        ;;
    integrated)
        TEST_FILE="integrated-load-test.js"
        k6 run \
            -e AUTH_URL="$AUTH_URL" \
            -e PRODUCT_URL="$PRODUCT_URL" \
            -e ORDER_URL="$ORDER_URL" \
            -e UI_URL="$UI_URL" \
            -e WARMUP_DURATION="$WARMUP_DURATION" \
            -e TEST_DURATION="$TEST_DURATION" \
            "$TEST_FILE"
        ;;
    all)
        log_info "Running all tests sequentially..."
        $0 auth
        sleep 30
        $0 product
        sleep 30
        $0 order
        sleep 30
        $0 integrated
        ;;
    *)
        log_error "Unknown test: $TEST_NAME"
        log_info "Available tests: auth, product, order, ui, integrated, all"
        exit 1
        ;;
esac

# ========================================================================
# Show HPA Status After Test
# ========================================================================

log_step "HPA status after test:"
kubectl get hpa -n $NAMESPACE 2>/dev/null || log_info "No HPA resources found"

log_step "Pod status:"
kubectl get pods -n $NAMESPACE -l app.kubernetes.io/part-of=velure

# ========================================================================
# Cleanup
# ========================================================================

if [ -z "$INGRESS_HOST" ]; then
    log_info "Cleaning up port-forwards..."
    pkill -f "kubectl port-forward" 2>/dev/null || true
fi

log_info "✅ Test completed!"
log_info ""
log_info "💡 Tips:"
log_info "  - Run './monitor-scaling.sh' in another terminal to watch scaling in real-time"
log_info "  - Check Grafana dashboard: http://localhost:3000"
log_info "  - View HPA: kubectl get hpa -n $NAMESPACE -w"

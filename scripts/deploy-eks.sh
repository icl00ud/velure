#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AWS_REGION="${AWS_REGION:-us-east-1}"
CLUSTER_NAME="${CLUSTER_NAME:-velure-production}"
TERRAFORM_DIR="infrastructure/terraform"

echo -e "${GREEN}=== Velure EKS Deployment Script ===${NC}"

# Change to project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"
echo "Working directory: $PROJECT_ROOT"

# Function to print step
step() {
    echo -e "\n${YELLOW}>>> $1${NC}"
}

# Function to check command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        exit 1
    fi
}

# Function to check if deployment exists and is ready
is_deployment_ready() {
    local name=$1
    local namespace=$2
    kubectl get deployment "$name" -n "$namespace" &>/dev/null && \
    kubectl get deployment "$name" -n "$namespace" -o jsonpath='{.status.readyReplicas}' 2>/dev/null | grep -q '[0-9]'
}

# Function to check if Helm release exists
helm_release_exists() {
    local name=$1
    local namespace=$2
    helm status "$name" -n "$namespace" &>/dev/null
}

# Function to check if namespace exists
namespace_exists() {
    kubectl get namespace "$1" &>/dev/null
}

# Check prerequisites
step "Checking prerequisites..."
check_command aws
check_command kubectl
check_command helm
check_command terraform

# Step 1: Configure kubectl
step "Configuring kubectl for EKS cluster..."
aws eks update-kubeconfig --region $AWS_REGION --name $CLUSTER_NAME

echo "Waiting for nodes to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s || true
kubectl get nodes

# Step 2: Add Helm repositories
step "Adding Helm repositories..."
helm repo add eks https://aws.github.io/eks-charts || true
helm repo add external-secrets https://charts.external-secrets.io || true
helm repo update

# Step 3: Install AWS Load Balancer Controller
step "Installing AWS Load Balancer Controller..."
if helm_release_exists "aws-load-balancer-controller" "kube-system"; then
    echo -e "${BLUE}AWS Load Balancer Controller already installed, skipping...${NC}"
else
    LB_ROLE_ARN=$(terraform -chdir=$TERRAFORM_DIR output -raw alb_controller_role_arn 2>/dev/null || echo "")
    VPC_ID=$(terraform -chdir=$TERRAFORM_DIR output -raw vpc_id 2>/dev/null || echo "")

    if [ -z "$LB_ROLE_ARN" ] || [ -z "$VPC_ID" ]; then
        echo -e "${YELLOW}Warning: Could not get Load Balancer Controller role ARN or VPC ID from Terraform${NC}"
        echo "Skipping AWS Load Balancer Controller installation"
    else
        helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
            -n kube-system \
            --set clusterName=$CLUSTER_NAME \
            --set serviceAccount.create=true \
            --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=$LB_ROLE_ARN \
            --set region=$AWS_REGION \
            --set vpcId=$VPC_ID \
            --wait --timeout 5m
    fi
fi

# Step 3.5: Install Metrics Server (required for HPA)
step "Installing Metrics Server..."
if is_deployment_ready "metrics-server" "kube-system"; then
    echo -e "${BLUE}Metrics Server already installed and ready, skipping...${NC}"
else
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    echo "Waiting for Metrics Server to be ready..."
    kubectl wait --for=condition=Available deployment/metrics-server -n kube-system --timeout=120s || true
fi

# Step 4: Install External Secrets Operator
step "Installing External Secrets Operator..."
if helm_release_exists "external-secrets" "external-secrets"; then
    echo -e "${BLUE}External Secrets Operator already installed, skipping...${NC}"
else
    helm upgrade --install external-secrets external-secrets/external-secrets \
        -n external-secrets --create-namespace \
        --wait --timeout 5m
fi

# Wait for External Secrets Operator to be ready
echo "Waiting for External Secrets Operator to be ready..."
kubectl wait --for=condition=Available deployment/external-secrets -n external-secrets --timeout=120s || true

# Wait for CRDs to be established
echo "Waiting for External Secrets CRDs to be ready..."
kubectl wait --for=condition=Established crd/externalsecrets.external-secrets.io --timeout=60s || true
kubectl wait --for=condition=Established crd/clustersecretstores.external-secrets.io --timeout=60s || true

# Create service account for External Secrets with IRSA
EXTERNAL_SECRETS_ROLE_ARN=$(terraform -chdir=$TERRAFORM_DIR output -raw external_secrets_role_arn 2>/dev/null || echo "")
if [ -n "$EXTERNAL_SECRETS_ROLE_ARN" ]; then
    echo "Creating External Secrets service account with IAM role..."
    kubectl create serviceaccount external-secrets-sa -n external-secrets --dry-run=client -o yaml | kubectl apply -f -
    kubectl annotate serviceaccount external-secrets-sa -n external-secrets \
        eks.amazonaws.com/role-arn=$EXTERNAL_SECRETS_ROLE_ARN --overwrite
else
    echo -e "${YELLOW}Warning: Could not get External Secrets role ARN from Terraform${NC}"
fi

# Step 4.5: Add VPC CIDR rule to RDS security group
step "Configuring RDS security group access..."
RDS_SG_ID=$(terraform -chdir=$TERRAFORM_DIR output -raw rds_security_group_id 2>/dev/null || echo "")
VPC_CIDR=$(terraform -chdir=$TERRAFORM_DIR output -raw vpc_cidr 2>/dev/null || echo "")
if [ -n "$RDS_SG_ID" ] && [ -n "$VPC_CIDR" ]; then
    echo "Adding VPC CIDR rule to RDS security group..."
    aws ec2 authorize-security-group-ingress \
        --group-id $RDS_SG_ID \
        --protocol tcp \
        --port 5432 \
        --cidr "$VPC_CIDR" \
        --region $AWS_REGION 2>/dev/null || echo "Rule already exists or error (continuing...)"
else
    echo -e "${YELLOW}Warning: Could not get RDS security group or VPC CIDR${NC}"
fi

# Step 5: Create namespaces
step "Creating namespaces..."
kubectl create ns authentication --dry-run=client -o yaml | kubectl apply -f -
kubectl create ns product --dry-run=client -o yaml | kubectl apply -f -
kubectl create ns order --dry-run=client -o yaml | kubectl apply -f -
kubectl create ns frontend --dry-run=client -o yaml | kubectl apply -f -

# Step 6: Apply External Secrets
step "Applying External Secrets configurations..."
if [ -d "infrastructure/kubernetes/external-secrets" ]; then
    # Clean up old ExternalSecrets that may have different names
    echo "Cleaning up old ExternalSecrets..."
    kubectl delete externalsecret -n authentication --all 2>/dev/null || true
    kubectl delete externalsecret -n order --all 2>/dev/null || true

    kubectl apply -f infrastructure/kubernetes/external-secrets/
    echo "Waiting for secrets to sync..."
    sleep 10

    # Force refresh all ExternalSecrets
    echo "Forcing refresh of ExternalSecrets..."
    for ns in authentication product order; do
        for es in $(kubectl get externalsecrets -n $ns -o name 2>/dev/null); do
            kubectl annotate $es -n $ns force-sync=$(date +%s) --overwrite 2>/dev/null || true
        done
    done

    # Wait for secrets to be ready
    echo "Waiting for secrets to sync..."
    sleep 15
    kubectl get externalsecrets -A

    # Check if any secrets failed
    FAILED=$(kubectl get externalsecrets -A -o jsonpath='{.items[?(@.status.conditions[0].reason=="SecretSyncedError")].metadata.name}' 2>/dev/null)
    if [ -n "$FAILED" ]; then
        echo -e "${YELLOW}Warning: Some ExternalSecrets failed to sync: $FAILED${NC}"
        echo "Continuing anyway..."
    fi
else
    echo -e "${YELLOW}Warning: External Secrets directory not found${NC}"
fi

# Step 6.5: Create Redis secrets (not from Secrets Manager)
step "Creating Redis secrets..."
kubectl create secret generic redis \
    --from-literal=redis-password="" \
    -n product --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic redis \
    --from-literal=redis-password="" \
    -n order --dry-run=client -o yaml | kubectl apply -f -

# Step 7: Deploy services
step "Deploying Velure services..."

deploy_service() {
    local name=$1
    local chart=$2
    local namespace=$3

    echo "Deploying $name..."

    if [ -d "$chart" ]; then
        helm upgrade --install $name $chart \
            -n $namespace \
            --wait --timeout 5m
    else
        echo -e "${RED}Error: Chart not found at $chart${NC}"
        return 1
    fi
}

deploy_service "velure-auth" "./infrastructure/kubernetes/charts/velure-auth" "authentication"
deploy_service "velure-product" "./infrastructure/kubernetes/charts/velure-product" "product"
deploy_service "velure-publish-order" "./infrastructure/kubernetes/charts/velure-publish-order" "order"
deploy_service "velure-process-order" "./infrastructure/kubernetes/charts/velure-process-order" "order"
deploy_service "velure-ui" "./infrastructure/kubernetes/charts/velure-ui" "frontend"

# Step 8: Deploy Observability Stack
step "Deploying Observability Stack (Prometheus + Grafana + Loki)..."

# Create monitoring namespace
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# Add Helm repositories for monitoring
echo "Adding monitoring Helm repositories..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
helm repo add grafana https://grafana.github.io/helm-charts || true
helm repo update

# Create Grafana dashboards ConfigMap from JSON files
echo "Creating Grafana dashboards ConfigMap..."
kubectl create configmap velure-grafana-dashboards \
    --from-file=infrastructure/kubernetes/monitoring/dashboards/ \
    -n monitoring \
    --dry-run=client -o yaml | kubectl apply -f -

# Install kube-prometheus-stack (Prometheus + Grafana + AlertManager)
if helm_release_exists "kube-prometheus-stack" "monitoring"; then
    echo -e "${BLUE}kube-prometheus-stack already installed, skipping...${NC}"
else
    echo "Installing kube-prometheus-stack..."
    helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
        -f infrastructure/kubernetes/monitoring/kube-prometheus-stack-values.yaml \
        -n monitoring \
        --wait --timeout 15m
fi

# Install Loki stack (for logs)
if helm_release_exists "loki" "monitoring"; then
    echo -e "${BLUE}Loki stack already installed, skipping...${NC}"
else
    echo "Installing Loki stack..."
    helm upgrade --install loki grafana/loki-stack \
        -f infrastructure/kubernetes/monitoring/loki-stack-values.yaml \
        -n monitoring \
        --wait --timeout 5m || echo "Loki installation skipped or failed"
fi

# Apply Grafana Ingress
echo "Applying Grafana Ingress..."
kubectl apply -f infrastructure/kubernetes/monitoring/grafana-ingress.yaml

# Wait for Prometheus to be ready
echo "Waiting for Prometheus to be ready..."
kubectl wait --for=condition=Ready pods -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s || true

# Apply ServiceMonitors
echo "Applying ServiceMonitors..."
if [ -d "infrastructure/kubernetes/monitoring/servicemonitors" ]; then
    kubectl apply -f infrastructure/kubernetes/monitoring/servicemonitors/ || echo "ServiceMonitors not yet available"
fi

# Apply Alert and Recording Rules
echo "Applying Prometheus rules..."
kubectl apply -f infrastructure/kubernetes/monitoring/alert-rules.yaml 2>/dev/null || echo "Alert rules not yet available"
kubectl apply -f infrastructure/kubernetes/monitoring/recording-rules.yaml 2>/dev/null || echo "Recording rules not yet available"

echo -e "${GREEN}Observability stack deployed successfully!${NC}"
echo "Grafana will be available via ALB in a few minutes"

# Get Grafana admin password
GRAFANA_PASSWORD=$(kubectl get secret -n monitoring kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" 2>/dev/null | base64 --decode 2>/dev/null || echo "admin")
echo "Grafana credentials: admin / $GRAFANA_PASSWORD"

# Step 9: Verify deployment
step "Verifying deployment..."
echo -e "\n${GREEN}=== Pods ===${NC}"
kubectl get pods -A -l app.kubernetes.io/managed-by=Helm

echo -e "\n${GREEN}=== Services ===${NC}"
kubectl get svc -A | grep velure

echo -e "\n${GREEN}=== Ingress ===${NC}"
kubectl get ingress -A

# Step 10: Get Load Balancer URLs
step "Getting application URLs..."

echo -e "\n${GREEN}=== Application URLs ===${NC}"
echo "Velure Services:"
kubectl get ingress -n frontend velure-ui -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null && echo " (Frontend)" || echo "Frontend URL not yet available"
kubectl get ingress -n authentication -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}' 2>/dev/null && echo " (Auth API)" || true
kubectl get ingress -n order -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}' 2>/dev/null && echo " (Product API)" || true

echo -e "\n${GREEN}Observability:${NC}"
GRAFANA_URL=$(kubectl get ingress -n monitoring grafana -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
if [ -n "$GRAFANA_URL" ]; then
    echo -e "${GREEN}Grafana: http://$GRAFANA_URL${NC}"
    echo "  Username: admin"
    echo "  Password: $GRAFANA_PASSWORD"
else
    echo "Grafana URL not yet available. Check with:"
    echo "  kubectl get ingress -n monitoring grafana"
    echo "  Or use port-forward: kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"
fi

echo -e "\n${GREEN}=== Deployment Complete ===${NC}"
echo "Useful commands:"
echo "  kubectl get pods -A                    # List all pods"
echo "  kubectl logs -f <pod> -n <namespace>   # View pod logs"
echo "  kubectl describe pod <pod> -n <ns>     # Debug pod issues"
echo ""
echo "Monitoring commands:"
echo "  kubectl get ingress -A                 # List all ingresses"
echo "  kubectl get servicemonitors -A         # List Prometheus targets"
echo "  kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090  # Access Prometheus"

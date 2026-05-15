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

# Function to retry kubectl commands on transient failures
retry_kubectl() {
    local max_attempts=5
    local timeout=2
    local attempt=1
    local exit_code=0

    while [ $attempt -le $max_attempts ]; do
        if "$@"; then
            return 0
        else
            exit_code=$?
            if [ $attempt -lt $max_attempts ]; then
                echo -e "${YELLOW}Command failed (attempt $attempt/$max_attempts). Retrying in ${timeout}s...${NC}" >&2
                sleep $timeout
                timeout=$((timeout * 2))  # Exponential backoff
                attempt=$((attempt + 1))
            fi
        fi
    done

    echo -e "${RED}Command failed after $max_attempts attempts${NC}" >&2
    return $exit_code
}

# Function to clean up stuck Helm releases and pods
cleanup_stuck_release() {
    local name=$1
    local namespace=$2
    
    echo "Checking for stuck release: $name in namespace $namespace..."
    
    # Check if there are pods in problematic states (check both label patterns)
    STUCK_PODS=$(kubectl get pods -n "$namespace" -l "app.kubernetes.io/instance=$name" -o jsonpath='{.items[?(@.status.phase!="Running")].metadata.name}' 2>/dev/null)
    if [ -z "$STUCK_PODS" ]; then
        STUCK_PODS=$(kubectl get pods -n "$namespace" -l "app=$name" -o jsonpath='{.items[?(@.status.phase!="Running")].metadata.name}' 2>/dev/null)
    fi
    
    # Check if helm release is in a bad state
    RELEASE_STATUS=$(helm status "$name" -n "$namespace" -o json 2>/dev/null | jq -r '.info.status' 2>/dev/null || echo "")
    
    if [ "$RELEASE_STATUS" = "failed" ] || [ "$RELEASE_STATUS" = "pending-install" ] || [ "$RELEASE_STATUS" = "pending-upgrade" ]; then
        echo -e "${YELLOW}Helm release $name is in '$RELEASE_STATUS' state, cleaning up...${NC}"
        
        # Uninstall the failed release
        helm uninstall "$name" -n "$namespace" --wait 2>/dev/null || true
        
        # Delete any remaining pods
        kubectl delete pods -n "$namespace" -l "app.kubernetes.io/instance=$name" --force --grace-period=0 2>/dev/null || true
        kubectl delete pods -n "$namespace" -l "app=$name" --force --grace-period=0 2>/dev/null || true
        
        # Wait for cleanup
        sleep 5
    elif [ -n "$STUCK_PODS" ]; then
        echo -e "${YELLOW}Found stuck pods for $name: $STUCK_PODS${NC}"
        # Just delete the stuck pods, Helm will recreate them
        kubectl delete pods -n "$namespace" -l "app.kubernetes.io/instance=$name" --force --grace-period=0 2>/dev/null || true
        kubectl delete pods -n "$namespace" -l "app=$name" --force --grace-period=0 2>/dev/null || true
        sleep 3
    fi
}

# Function to adopt existing secrets into Helm release
adopt_secrets_for_helm() {
    local release_name=$1
    local namespace=$2
    
    echo "Adopting existing secrets for Helm release: $release_name..."
    
    # Get all secrets in namespace that might conflict
    for secret in $(kubectl get secrets -n "$namespace" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
        # Check if secret has Helm labels already
        HAS_HELM_LABEL=$(kubectl get secret "$secret" -n "$namespace" -o jsonpath='{.metadata.labels.app\.kubernetes\.io/managed-by}' 2>/dev/null || echo "")
        
        if [ -z "$HAS_HELM_LABEL" ]; then
            # Add Helm ownership labels and annotations
            kubectl label secret "$secret" -n "$namespace" \
                "app.kubernetes.io/managed-by=Helm" \
                --overwrite 2>/dev/null || true
            kubectl annotate secret "$secret" -n "$namespace" \
                "meta.helm.sh/release-name=$release_name" \
                "meta.helm.sh/release-namespace=$namespace" \
                --overwrite 2>/dev/null || true
        fi
    done
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
helm repo add bitnami https://charts.bitnami.com/bitnami || true
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
            --wait --timeout 1m
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

# Step 3.6: Create gp3 StorageClass (required for datastores)
step "Creating gp3 StorageClass..."
if kubectl get storageclass gp3 &>/dev/null; then
    echo -e "${BLUE}StorageClass gp3 already exists, skipping...${NC}"
else
    cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp3
  fsType: ext4
  encrypted: "true"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
EOF
    echo -e "${GREEN}StorageClass gp3 created successfully${NC}"
fi

# Step 4: Install External Secrets Operator
step "Installing External Secrets Operator..."
if helm_release_exists "external-secrets" "external-secrets"; then
    echo -e "${BLUE}External Secrets Operator already installed, skipping...${NC}"
else
    helm upgrade --install external-secrets external-secrets/external-secrets \
        -n external-secrets --create-namespace \
        --wait --timeout 1m
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
kubectl create ns datastores --dry-run=client -o yaml | kubectl apply -f -

# Step 6: Apply External Secrets
step "Applying External Secrets configurations..."
if [ -d "infrastructure/kubernetes/external-secrets" ]; then
    # Apply ExternalSecrets (idempotent - no need to delete first)
    kubectl apply -f infrastructure/kubernetes/external-secrets/
    
    echo "Waiting for all ExternalSecrets to sync..."
    
    # Wait for each secret to be ready with proper timeout
    wait_for_secret() {
        local ns=$1
        local secret_name=$2
        local max_attempts=30
        local attempt=0
        
        echo "  Waiting for secret $secret_name in $ns..."
        while [ $attempt -lt $max_attempts ]; do
            if kubectl get secret "$secret_name" -n "$ns" &>/dev/null; then
                echo "  ✓ Secret $secret_name ready"
                return 0
            fi
            attempt=$((attempt + 1))
            sleep 2
        done
        echo -e "${RED}  ✗ Secret $secret_name not ready after ${max_attempts} attempts${NC}"
        return 1
    }
    
    # Wait for all required secrets
    SECRETS_OK=true
    wait_for_secret "authentication" "velure-auth-postgres" || SECRETS_OK=false
    wait_for_secret "authentication" "velure-auth-jwt" || SECRETS_OK=false
    wait_for_secret "authentication" "velure-auth-session" || SECRETS_OK=false
    wait_for_secret "product" "velure-product-secret" || SECRETS_OK=false
    wait_for_secret "order" "order-database" || SECRETS_OK=false
    wait_for_secret "order" "rabbitmq-conn" || SECRETS_OK=false
    wait_for_secret "order" "order-jwt" || SECRETS_OK=false
    
    if [ "$SECRETS_OK" = false ]; then
        echo -e "${RED}Error: Some secrets failed to sync. Check ExternalSecrets status:${NC}"
        kubectl get externalsecrets -A
        echo -e "\nDescribe failed ExternalSecrets for details:"
        echo "  kubectl describe externalsecret -n <namespace> <name>"
        exit 1
    fi
    
    echo -e "${GREEN}All secrets synced successfully${NC}"
    kubectl get externalsecrets -A
else
    echo -e "${YELLOW}Warning: External Secrets directory not found${NC}"
fi

# Step 6.5: Deploy Observability Stack
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
        --wait --timeout 10m
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
# Step 7: Deploy datastores (Redis for auth-service)
step "Deploying datastores (Redis)..."
if helm_release_exists "velure-datastores" "datastores"; then
    echo -e "${BLUE}Datastores already installed, upgrading...${NC}"
fi

# Build Helm dependencies first (mongodb, redis, rabbitmq subchart dependencies)
echo "Building Helm chart dependencies..."
DATASTORES_CHART="./infrastructure/kubernetes/charts/velure-datastores"

# Remove stale lock file and rebuild dependencies
rm -f "${DATASTORES_CHART}/Chart.lock" 2>/dev/null || true
helm dependency update "${DATASTORES_CHART}" || {
    echo -e "${YELLOW}Warning: Failed to update dependencies, trying build...${NC}"
    helm dependency build "${DATASTORES_CHART}" || {
        echo -e "${RED}Error: Failed to build Helm dependencies${NC}"
        echo "Try manually: helm dependency update ${DATASTORES_CHART}"
    }
}

# Deploy Redis only (disable MongoDB and RabbitMQ since they use managed services)
helm upgrade --install velure-datastores "${DATASTORES_CHART}" \
    -n datastores \
    --set mongodb.enabled=false \
    --set rabbitmq.enabled=false \
    --set redis.enabled=true \
    --set redis.auth.password="redis_secure_password" \
    --wait --timeout 3m || {
        echo -e "${YELLOW}Warning: Redis deployment may have issues, checking status...${NC}"
        kubectl get pods -n datastores
    }

# Create Redis secret for auth-service
echo "Creating Redis secrets for services..."
retry_kubectl bash -c "kubectl create secret generic velure-auth-redis \
    --from-literal=redis-password='redis_secure_password' \
    -n authentication --dry-run=client -o yaml | kubectl apply -f -"

retry_kubectl bash -c "kubectl create secret generic redis \
    --from-literal=redis-password='redis_secure_password' \
    -n product --dry-run=client -o yaml | kubectl apply -f -"

# Wait for Redis to be ready before deploying services
echo "Waiting for Redis to be ready..."
REDIS_READY=false
for i in {1..30}; do
    if kubectl wait --for=condition=Ready pods -l app.kubernetes.io/name=redis -n datastores --timeout=10s 2>/dev/null; then
        REDIS_READY=true
        echo -e "${GREEN}Redis is ready!${NC}"
        break
    fi
    echo "Waiting for Redis... attempt $i/30"
    sleep 5
done

if [ "$REDIS_READY" = false ]; then
    echo -e "${YELLOW}Warning: Redis not ready after waiting. Services may fail to start.${NC}"
    kubectl get pods -n datastores
fi

# Step 7.5: Bootstrap RabbitMQ queues and exchanges
step "Bootstrapping RabbitMQ queues..."
if kubectl get job rabbitmq-bootstrap -n order &>/dev/null; then
    echo -e "${BLUE}RabbitMQ bootstrap job already exists, deleting old job...${NC}"
    kubectl delete job rabbitmq-bootstrap -n order --ignore-not-found=true
    sleep 2
fi

echo "Creating RabbitMQ bootstrap job..."
kubectl apply -f infrastructure/kubernetes/jobs/rabbitmq-bootstrap-job.yaml

echo "Waiting for RabbitMQ bootstrap to complete..."
if kubectl wait --for=condition=complete job/rabbitmq-bootstrap -n order --timeout=120s 2>/dev/null; then
    echo -e "${GREEN}✓ RabbitMQ queues and exchanges created successfully${NC}"
    kubectl logs job/rabbitmq-bootstrap -n order | tail -20
else
    echo -e "${YELLOW}Warning: RabbitMQ bootstrap may have issues${NC}"
    kubectl describe job rabbitmq-bootstrap -n order
    kubectl logs job/rabbitmq-bootstrap -n order --tail=50 || true
fi

# Step 8: Deploy services
step "Deploying Velure services..."

deploy_service() {
    local name=$1
    local chart=$2
    local namespace=$3
    local timeout=${4:-1m}  # Default 1 minute timeout

    echo -e "${BLUE}Deploying $name...${NC}"

    if [ ! -d "$chart" ]; then
        echo -e "${RED}Error: Chart not found at $chart${NC}"
        return 1
    fi
    
    # Clean up any stuck releases before deploying
    cleanup_stuck_release "$name" "$namespace"
    
    # Adopt existing secrets so Helm can manage them
    adopt_secrets_for_helm "$name" "$namespace"
    
    # Deploy with increased timeout and atomic flag for rollback on failure
    if ! helm upgrade --install "$name" "$chart" \
        -n "$namespace" \
        --set image.tag=latest \
        --set image.pullPolicy=Always \
        --atomic \
        --timeout "$timeout"; then
        
        echo -e "${RED}Helm deployment failed for $name. Debugging...${NC}"
        echo "Pod status:"
        kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=$name" -o wide 2>/dev/null || \
            kubectl get pods -n "$namespace" -l "app=$name" -o wide 2>/dev/null || true
        echo "Pod events:"
        kubectl get events -n "$namespace" --sort-by='.lastTimestamp' | tail -20
        echo "Pod logs (last 50 lines):"
        POD=$(kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=$name" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
        if [ -n "$POD" ]; then
            kubectl logs "$POD" -n "$namespace" --tail=50 2>/dev/null || true
        fi
        return 1
    fi
    
    echo -e "${GREEN}✓ $name deployed successfully${NC}"
}

deploy_service "velure-auth" "./infrastructure/kubernetes/charts/velure-auth" "authentication" "3m"
deploy_service "velure-product" "./infrastructure/kubernetes/charts/velure-product" "product" "3m"
deploy_service "velure-publish-order" "./infrastructure/kubernetes/charts/velure-publish-order" "order" "3m"
deploy_service "velure-process-order" "./infrastructure/kubernetes/charts/velure-process-order" "order" "3m"
deploy_service "velure-ui" "./infrastructure/kubernetes/charts/velure-ui" "frontend" "3m"

# Step 8: Deploy Observability Stack
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

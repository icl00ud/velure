#!/usr/bin/env bash
set -euo pipefail

# ===========================================================================================
# Velure Bootstrap Script
# ===========================================================================================
# Script de automação completa para subir infraestrutura AWS + Kubernetes + Aplicação
# 
# Uso: ./scripts/bootstrap.sh [OPTIONS]
# 
# Opções:
#   --skip-terraform     Pula criação da infraestrutura AWS (usa existente)
#   --skip-secrets       Pula criação de secrets no AWS Secrets Manager
#   --skip-k8s-setup     Pula instalação de controllers/operators
#   --skip-datastores    Pula deploy dos datastores
#   --skip-services      Pula deploy dos microserviços
#   --destroy            Destroi toda a infraestrutura (PERIGOSO!)
# ===========================================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações padrão
PROJECT_NAME="velure"
ENVIRONMENT="prod"
AWS_REGION="us-east-1"
CLUSTER_NAME="${PROJECT_NAME}-${ENVIRONMENT}"

SKIP_TERRAFORM=false
SKIP_SECRETS=false
SKIP_K8S_SETUP=false
SKIP_DATASTORES=false
SKIP_SERVICES=false
DESTROY_MODE=false

# ===========================================================================================
# Funções auxiliares
# ===========================================================================================

log_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

check_prerequisites() {
    log_info "Verificando pré-requisitos..."
    
    local missing=()
    
    command -v aws >/dev/null 2>&1 || missing+=("aws-cli")
    command -v terraform >/dev/null 2>&1 || missing+=("terraform")
    command -v kubectl >/dev/null 2>&1 || missing+=("kubectl")
    command -v helm >/dev/null 2>&1 || missing+=("helm")
    command -v jq >/dev/null 2>&1 || missing+=("jq")
    
    if [ ${#missing[@]} -ne 0 ]; then
        log_error "Ferramentas faltando: ${missing[*]}"
        log_info "Instale com: brew install ${missing[*]}"
        exit 1
    fi
    
    # Verificar credenciais AWS
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        log_error "Credenciais AWS não configuradas"
        log_info "Execute: aws configure"
        exit 1
    fi
    
    log_success "Pré-requisitos OK"
}

generate_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

# ===========================================================================================
# Gestão de Secrets
# ===========================================================================================

create_secrets_in_aws() {
    log_info "Criando secrets no AWS Secrets Manager..."
    
    local secrets_prefix="${PROJECT_NAME}/${ENVIRONMENT}"
    
    # Gerar senhas seguras
    local rds_auth_password=$(generate_password)
    local rds_orders_password=$(generate_password)
    local rabbitmq_password=$(generate_password)
    local jwt_secret=$(generate_password)
    local jwt_refresh_secret=$(generate_password)
    local session_secret=$(generate_password)
    
    # RDS Auth Database
    aws secretsmanager create-secret \
        --name "${secrets_prefix}/rds-auth" \
        --description "RDS Auth Service Database Credentials" \
        --secret-string "{\"username\":\"postgres\",\"password\":\"${rds_auth_password}\",\"dbname\":\"velure_auth\"}" \
        --region "${AWS_REGION}" 2>/dev/null || \
    aws secretsmanager update-secret \
        --secret-id "${secrets_prefix}/rds-auth" \
        --secret-string "{\"username\":\"postgres\",\"password\":\"${rds_auth_password}\",\"dbname\":\"velure_auth\"}" \
        --region "${AWS_REGION}"
    
    # RDS Orders Database
    aws secretsmanager create-secret \
        --name "${secrets_prefix}/rds-orders" \
        --description "RDS Orders Service Database Credentials" \
        --secret-string "{\"username\":\"postgres\",\"password\":\"${rds_orders_password}\",\"dbname\":\"velure_orders\"}" \
        --region "${AWS_REGION}" 2>/dev/null || \
    aws secretsmanager update-secret \
        --secret-id "${secrets_prefix}/rds-orders" \
        --secret-string "{\"username\":\"postgres\",\"password\":\"${rds_orders_password}\",\"dbname\":\"velure_orders\"}" \
        --region "${AWS_REGION}"
    
    # RabbitMQ
    aws secretsmanager create-secret \
        --name "${secrets_prefix}/rabbitmq" \
        --description "RabbitMQ Admin Credentials" \
        --secret-string "{\"username\":\"admin\",\"password\":\"${rabbitmq_password}\"}" \
        --region "${AWS_REGION}" 2>/dev/null || \
    aws secretsmanager update-secret \
        --secret-id "${secrets_prefix}/rabbitmq" \
        --secret-string "{\"username\":\"admin\",\"password\":\"${rabbitmq_password}\"}" \
        --region "${AWS_REGION}"
    
    # JWT Secrets
    aws secretsmanager create-secret \
        --name "${secrets_prefix}/jwt" \
        --description "JWT Secrets for Auth Service" \
        --secret-string "{\"secret\":\"${jwt_secret}\",\"refreshSecret\":\"${jwt_refresh_secret}\",\"expiresIn\":\"1h\",\"refreshExpiresIn\":\"7d\"}" \
        --region "${AWS_REGION}" 2>/dev/null || \
    aws secretsmanager update-secret \
        --secret-id "${secrets_prefix}/jwt" \
        --secret-string "{\"secret\":\"${jwt_secret}\",\"refreshSecret\":\"${jwt_refresh_secret}\",\"expiresIn\":\"1h\",\"refreshExpiresIn\":\"7d\"}" \
        --region "${AWS_REGION}"
    
    # Session Secret
    aws secretsmanager create-secret \
        --name "${secrets_prefix}/session" \
        --description "Session Secret for Auth Service" \
        --secret-string "{\"secret\":\"${session_secret}\",\"expiresIn\":\"86400000\"}" \
        --region "${AWS_REGION}" 2>/dev/null || \
    aws secretsmanager update-secret \
        --secret-id "${secrets_prefix}/session" \
        --secret-string "{\"secret\":\"${session_secret}\",\"expiresIn\":\"86400000\"}" \
        --region "${AWS_REGION}"
    
    # MongoDB Atlas (ajuste se necessário)
    log_warning "MongoDB: Configure manualmente no MongoDB Atlas e adicione a connection string em ${secrets_prefix}/mongodb"
    
    log_success "Secrets criados no AWS Secrets Manager"
    log_info "Prefixo: ${secrets_prefix}/"
}

get_secret_from_aws() {
    local secret_name=$1
    local key=$2
    aws secretsmanager get-secret-value \
        --secret-id "${PROJECT_NAME}/${ENVIRONMENT}/${secret_name}" \
        --region "${AWS_REGION}" \
        --query 'SecretString' \
        --output text | jq -r ".${key}"
}

# ===========================================================================================
# Infraestrutura AWS (Terraform)
# ===========================================================================================

deploy_terraform() {
    log_info "Deploying infraestrutura AWS com Terraform..."
    
    cd infrastructure/terraform
    
    # Obter senhas do Secrets Manager
    local rds_auth_password=$(get_secret_from_aws "rds-auth" "password")
    local rds_orders_password=$(get_secret_from_aws "rds-orders" "password")
    local rabbitmq_password=$(get_secret_from_aws "rabbitmq" "password")
    
    # Criar terraform.tfvars automaticamente
    cat > terraform.tfvars <<EOF
# Auto-generated by bootstrap.sh
aws_region = "${AWS_REGION}"
project_name = "${PROJECT_NAME}"
environment = "${ENVIRONMENT}"

# RDS Auth
rds_auth_password = "${rds_auth_password}"

# RDS Orders
rds_orders_password = "${rds_orders_password}"

# RabbitMQ
rabbitmq_admin_password = "${rabbitmq_password}"
EOF
    
    terraform init
    terraform plan -out=tfplan
    terraform apply tfplan
    
    # Salvar outputs
    terraform output -json > ../../.terraform-outputs.json
    
    cd ../..
    
    log_success "Infraestrutura AWS criada"
}

destroy_terraform() {
    log_warning "DESTRUINDO infraestrutura AWS..."
    cd infrastructure/terraform
    terraform destroy -auto-approve
    cd ../..
    log_success "Infraestrutura AWS destruída"
}

# ===========================================================================================
# Kubernetes Setup
# ===========================================================================================

configure_kubectl() {
    log_info "Configurando kubectl para EKS..."
    aws eks update-kubeconfig \
        --region "${AWS_REGION}" \
        --name "${CLUSTER_NAME}"
    log_success "kubectl configurado"
}

install_external_secrets_operator() {
    log_info "Instalando External Secrets Operator..."
    
    helm repo add external-secrets https://charts.external-secrets.io 2>/dev/null || true
    helm repo update
    
    helm upgrade --install external-secrets \
        external-secrets/external-secrets \
        -n external-secrets-system \
        --create-namespace \
        --set installCRDs=true \
        --wait
    
    # Obter AWS Account ID
    local aws_account_id=$(aws sts get-caller-identity --query Account --output text)
    local aws_region="${AWS_REGION}"
    
    # Criar IAM Role para External Secrets
    kubectl create namespace velure 2>/dev/null || true
    
    # Criar SecretStore
    cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: ${aws_region}
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets
            namespace: external-secrets-system
EOF
    
    log_success "External Secrets Operator instalado"
}

install_alb_controller() {
    log_info "Instalando AWS Load Balancer Controller..."
    
    local cluster_name="${CLUSTER_NAME}"
    local aws_account_id=$(aws sts get-caller-identity --query Account --output text)
    
    # Criar IAM Policy
    curl -o iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json
    
    aws iam create-policy \
        --policy-name AWSLoadBalancerControllerIAMPolicy \
        --policy-document file://iam_policy.json 2>/dev/null || true
    
    rm iam_policy.json
    
    # Criar IAM Service Account
    eksctl create iamserviceaccount \
        --cluster="${cluster_name}" \
        --namespace=kube-system \
        --name=aws-load-balancer-controller \
        --attach-policy-arn=arn:aws:iam::${aws_account_id}:policy/AWSLoadBalancerControllerIAMPolicy \
        --override-existing-serviceaccounts \
        --approve 2>/dev/null || true
    
    # Instalar via Helm
    helm repo add eks https://aws.github.io/eks-charts 2>/dev/null || true
    helm repo update
    
    helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
        -n kube-system \
        --set clusterName="${cluster_name}" \
        --set serviceAccount.create=false \
        --set serviceAccount.name=aws-load-balancer-controller \
        --wait
    
    log_success "ALB Controller instalado"
}

install_metrics_server() {
    log_info "Instalando Metrics Server..."
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    log_success "Metrics Server instalado"
}

setup_kubernetes() {
    configure_kubectl
    install_external_secrets_operator
    install_alb_controller
    install_metrics_server
}

# ===========================================================================================
# Deploy de Datastores
# ===========================================================================================

deploy_datastores() {
    log_info "Deploying datastores (MongoDB, Redis, RabbitMQ in-cluster)..."
    
    kubectl create namespace datastores 2>/dev/null || true
    
    helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
    helm repo update
    
    # Deploy velure-datastores chart com External Secrets
    helm upgrade --install velure-datastores \
        infrastructure/kubernetes/charts/velure-datastores \
        -n datastores \
        --create-namespace \
        --dependency-update \
        --wait \
        --timeout=10m
    
    log_success "Datastores deployados"
}

# ===========================================================================================
# Deploy de Microserviços
# ===========================================================================================

create_external_secrets() {
    log_info "Criando ExternalSecrets para os serviços..."
    
    local secrets_prefix="${PROJECT_NAME}/${ENVIRONMENT}"
    
    # Auth Service Secrets
    cat <<EOF | kubectl apply -f -
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: velure-auth-jwt
  namespace: velure
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: velure-auth-jwt
    creationPolicy: Owner
  data:
    - secretKey: secret
      remoteRef:
        key: ${secrets_prefix}/jwt
        property: secret
    - secretKey: expiresIn
      remoteRef:
        key: ${secrets_prefix}/jwt
        property: expiresIn
    - secretKey: refreshSecret
      remoteRef:
        key: ${secrets_prefix}/jwt
        property: refreshSecret
    - secretKey: refreshExpiresIn
      remoteRef:
        key: ${secrets_prefix}/jwt
        property: refreshExpiresIn
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: velure-auth-session
  namespace: velure
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: velure-auth-session
    creationPolicy: Owner
  data:
    - secretKey: secret
      remoteRef:
        key: ${secrets_prefix}/session
        property: secret
    - secretKey: expiresIn
      remoteRef:
        key: ${secrets_prefix}/session
        property: expiresIn
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: velure-auth-postgres
  namespace: velure
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: velure-auth-postgres
    creationPolicy: Owner
  dataFrom:
    - extract:
        key: ${secrets_prefix}/rds-auth
EOF
    
    log_success "ExternalSecrets criados"
}

deploy_services() {
    log_info "Deploying microserviços Velure..."
    
    kubectl create namespace velure 2>/dev/null || true
    
    create_external_secrets
    
    # Obter endpoints RDS e RabbitMQ do Terraform
    local rds_auth_endpoint=$(jq -r '.rds_auth_endpoint.value' .terraform-outputs.json)
    local rds_orders_endpoint=$(jq -r '.rds_orders_endpoint.value' .terraform-outputs.json)
    local amazonmq_endpoint=$(jq -r '.amazonmq_endpoint.value' .terraform-outputs.json)
    
    # Auth Service
    helm upgrade --install velure-auth \
        infrastructure/kubernetes/charts/velure-auth \
        -n velure \
        --set database.host="${rds_auth_endpoint}" \
        --wait
    
    # Product Service
    helm upgrade --install velure-product \
        infrastructure/kubernetes/charts/velure-product \
        -n velure \
        --wait
    
    # Publish Order Service
    helm upgrade --install velure-publish-order \
        infrastructure/kubernetes/charts/velure-publish-order \
        -n velure \
        --set database.host="${rds_orders_endpoint}" \
        --set rabbitmq.host="${amazonmq_endpoint}" \
        --wait
    
    # Process Order Service
    helm upgrade --install velure-process-order \
        infrastructure/kubernetes/charts/velure-process-order \
        -n velure \
        --set database.host="${rds_orders_endpoint}" \
        --set rabbitmq.host="${amazonmq_endpoint}" \
        --wait
    
    # UI Service
    helm upgrade --install velure-ui \
        infrastructure/kubernetes/charts/velure-ui \
        -n velure \
        --wait
    
    log_success "Microserviços deployados"
}

# ===========================================================================================
# Monitoramento
# ===========================================================================================

install_monitoring() {
    log_info "Instalando stack de monitoramento..."
    
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
    helm repo add grafana https://grafana.github.io/helm-charts 2>/dev/null || true
    helm repo update
    
    kubectl create namespace monitoring 2>/dev/null || true
    
    # Prometheus + Grafana
    helm upgrade --install kube-prometheus-stack \
        prometheus-community/kube-prometheus-stack \
        -n monitoring \
        --create-namespace \
        --wait
    
    log_success "Monitoramento instalado"
    log_info "Grafana: kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"
    log_info "Senha admin: kubectl get secret -n monitoring kube-prometheus-stack-grafana -o jsonpath='{.data.admin-password}' | base64 --decode"
}

# ===========================================================================================
# Destroy
# ===========================================================================================

destroy_all() {
    log_warning "DESTRUINDO TUDO!"
    
    read -p "Tem certeza? Digite 'DELETE' para confirmar: " confirm
    if [ "$confirm" != "DELETE" ]; then
        log_info "Cancelado"
        exit 0
    fi
    
    # Remover Kubernetes resources
    log_info "Removendo recursos Kubernetes..."
    helm uninstall velure-ui velure-auth velure-product velure-publish-order velure-process-order -n velure 2>/dev/null || true
    helm uninstall kube-prometheus-stack -n monitoring 2>/dev/null || true
    helm uninstall velure-datastores -n datastores 2>/dev/null || true
    helm uninstall external-secrets -n external-secrets-system 2>/dev/null || true
    helm uninstall aws-load-balancer-controller -n kube-system 2>/dev/null || true
    
    kubectl delete namespace velure monitoring datastores external-secrets-system 2>/dev/null || true
    kubectl delete pvc --all -n datastores 2>/dev/null || true
    
    # Destruir Terraform
    destroy_terraform
    
    log_success "Tudo destruído"
}

# ===========================================================================================
# Main
# ===========================================================================================

print_banner() {
    cat << "EOF"
╦  ╦┌─┐┬  ┬ ┬┬─┐┌─┐
╚╗╔╝├┤ │  │ │├┬┘├┤ 
 ╚╝ └─┘┴─┘└─┘┴└─└─┘
Bootstrap Automation Script
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-terraform) SKIP_TERRAFORM=true ;;
            --skip-secrets) SKIP_SECRETS=true ;;
            --skip-k8s-setup) SKIP_K8S_SETUP=true ;;
            --skip-datastores) SKIP_DATASTORES=true ;;
            --skip-services) SKIP_SERVICES=true ;;
            --destroy) DESTROY_MODE=true ;;
            -h|--help)
                echo "Uso: $0 [OPTIONS]"
                echo "Opções:"
                echo "  --skip-terraform     Pula criação da infraestrutura AWS"
                echo "  --skip-secrets       Pula criação de secrets"
                echo "  --skip-k8s-setup     Pula instalação de controllers"
                echo "  --skip-datastores    Pula deploy dos datastores"
                echo "  --skip-services      Pula deploy dos microserviços"
                echo "  --destroy            Destroi toda a infraestrutura"
                exit 0
                ;;
            *)
                log_error "Opção desconhecida: $1"
                exit 1
                ;;
        esac
        shift
    done
}

main() {
    print_banner
    parse_args "$@"
    
    if [ "$DESTROY_MODE" = true ]; then
        destroy_all
        exit 0
    fi
    
    check_prerequisites
    
    if [ "$SKIP_SECRETS" = false ]; then
        create_secrets_in_aws
    fi
    
    if [ "$SKIP_TERRAFORM" = false ]; then
        deploy_terraform
    fi
    
    if [ "$SKIP_K8S_SETUP" = false ]; then
        setup_kubernetes
    fi
    
    if [ "$SKIP_DATASTORES" = false ]; then
        deploy_datastores
    fi
    
    if [ "$SKIP_SERVICES" = false ]; then
        deploy_services
    fi
    
    install_monitoring
    
    log_success "🎉 Deploy completo!"
    echo ""
    log_info "Próximos passos:"
    echo "  1. Verificar pods: kubectl get pods -A"
    echo "  2. Ver ingress: kubectl get ingress -n velure"
    echo "  3. Grafana: make eks-grafana"
    echo "  4. Logs: kubectl logs -n velure -l app=velure-auth"
}

main "$@"

#!/bin/bash

# ======================================================================
# Script: 01-install-controllers.sh
# Descrição: Instala controladores essenciais no cluster EKS
#           - AWS Load Balancer Controller (ALB)
#           - metrics-server (para HPA)
# ======================================================================

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Funções helper
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar kubectl
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl não encontrado. Instale: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# Verificar helm
if ! command -v helm &> /dev/null; then
    log_error "helm não encontrado. Instale: https://helm.sh/docs/intro/install/"
    exit 1
fi

# Verificar conexão com cluster
log_info "Verificando conexão com cluster Kubernetes..."
if ! kubectl cluster-info &> /dev/null; then
    log_error "Não foi possível conectar ao cluster. Configure kubectl primeiro."
    log_info "Execute: aws eks update-kubeconfig --region us-east-1 --name velure-prod"
    exit 1
fi

CLUSTER_NAME=$(kubectl config current-context | awk -F'/' '{print $2}')
log_info "Conectado ao cluster: $CLUSTER_NAME"

# ======================================================================
# 1. Instalar metrics-server
# ======================================================================

log_info "=== Instalando metrics-server ==="

if kubectl get deployment metrics-server -n kube-system &> /dev/null; then
    log_warn "metrics-server já está instalado"
else
    log_info "Aplicando metrics-server..."
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

    log_info "Aguardando metrics-server ficar pronto..."
    kubectl wait --for=condition=available --timeout=300s deployment/metrics-server -n kube-system

    log_info "✓ metrics-server instalado com sucesso"
fi

# Verificar funcionamento
log_info "Testando metrics-server..."
sleep 10
if kubectl top nodes &> /dev/null; then
    log_info "✓ metrics-server está funcionando corretamente"
else
    log_warn "metrics-server ainda não retornou métricas (pode levar alguns minutos)"
fi

# ======================================================================
# 2. Instalar AWS Load Balancer Controller
# ======================================================================

log_info "=== Instalando AWS Load Balancer Controller ==="

# Obter Account ID e Region
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
AWS_REGION=$(aws configure get region || echo "us-east-1")

log_info "AWS Account ID: $AWS_ACCOUNT_ID"
log_info "AWS Region: $AWS_REGION"

# Criar IAM policy se não existir
POLICY_NAME="AWSLoadBalancerControllerIAMPolicy"
POLICY_ARN="arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}"

log_info "Verificando IAM policy..."
if aws iam get-policy --policy-arn "$POLICY_ARN" &> /dev/null; then
    log_warn "IAM Policy já existe: $POLICY_ARN"
else
    log_info "Criando IAM Policy..."

    # Download policy document
    curl -o /tmp/iam-policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.7.0/docs/install/iam_policy.json

    aws iam create-policy \
        --policy-name "$POLICY_NAME" \
        --policy-document file:///tmp/iam-policy.json

    log_info "✓ IAM Policy criada: $POLICY_ARN"
fi

# Criar ServiceAccount com IRSA
log_info "Criando ServiceAccount com IRSA..."

eksctl create iamserviceaccount \
    --cluster="$CLUSTER_NAME" \
    --namespace=kube-system \
    --name=aws-load-balancer-controller \
    --attach-policy-arn="$POLICY_ARN" \
    --override-existing-serviceaccounts \
    --region="$AWS_REGION" \
    --approve || log_warn "ServiceAccount pode já existir"

# Adicionar repositório Helm
log_info "Adicionando repositório Helm eks-charts..."
helm repo add eks https://aws.github.io/eks-charts
helm repo update

# Instalar ou atualizar AWS Load Balancer Controller
if helm list -n kube-system | grep -q aws-load-balancer-controller; then
    log_warn "AWS Load Balancer Controller já instalado, atualizando..."

    helm upgrade aws-load-balancer-controller eks/aws-load-balancer-controller \
        -n kube-system \
        --set clusterName="$CLUSTER_NAME" \
        --set serviceAccount.create=false \
        --set serviceAccount.name=aws-load-balancer-controller \
        --set region="$AWS_REGION" \
        --set vpcId=$(aws eks describe-cluster --name "$CLUSTER_NAME" --query "cluster.resourcesVpcConfig.vpcId" --output text)
else
    log_info "Instalando AWS Load Balancer Controller..."

    helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
        -n kube-system \
        --set clusterName="$CLUSTER_NAME" \
        --set serviceAccount.create=false \
        --set serviceAccount.name=aws-load-balancer-controller \
        --set region="$AWS_REGION" \
        --set vpcId=$(aws eks describe-cluster --name "$CLUSTER_NAME" --query "cluster.resourcesVpcConfig.vpcId" --output text)
fi

# Aguardar deployment ficar pronto
log_info "Aguardando AWS Load Balancer Controller ficar pronto..."
kubectl wait --for=condition=available --timeout=300s \
    deployment/aws-load-balancer-controller -n kube-system

log_info "✓ AWS Load Balancer Controller instalado com sucesso"

# ======================================================================
# Verificação Final
# ======================================================================

log_info "=== Verificação Final ==="

echo ""
log_info "Controllers instalados:"
kubectl get deployment -n kube-system | grep -E "(metrics-server|aws-load-balancer-controller)"

echo ""
log_info "ServiceAccounts:"
kubectl get serviceaccount -n kube-system | grep -E "(metrics-server|aws-load-balancer-controller)"

echo ""
log_info "✓ Instalação concluída com sucesso!"
log_info ""
log_info "Próximos passos:"
log_info "  1. Execute: ./02-install-datastores.sh"
log_info "  2. Execute: ./03-install-monitoring.sh"
log_info "  3. Execute: ./04-deploy-services.sh"

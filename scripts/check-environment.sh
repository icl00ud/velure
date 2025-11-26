#!/usr/bin/env bash

# Velure - Verificação de ambiente para deploy automatizado

echo "╦  ╦┌─┐┬  ┬ ┬┬─┐┌─┐"
echo "╚╗╔╝├┤ │  │ │├┬┘├┤ "
echo " ╚╝ └─┘┴─┘└─┘┴└─└─┘"
echo ""
echo "Verificando ambiente para deploy automatizado..."
echo ""

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

errors=0

# Verificar ferramentas
echo "🔍 Verificando ferramentas..."

check_tool() {
    if command -v $1 >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} $1 instalado: $(command -v $1)"
    else
        echo -e "  ${RED}✗${NC} $1 não encontrado"
        echo -e "     Instale com: brew install $2"
        ((errors++))
    fi
}

check_tool "aws" "awscli"
check_tool "terraform" "terraform"
check_tool "kubectl" "kubectl"
check_tool "helm" "helm"
check_tool "jq" "jq"

echo ""
echo "🔐 Verificando credenciais AWS..."

if aws sts get-caller-identity >/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} Credenciais AWS configuradas"
    aws sts get-caller-identity --output table
else
    echo -e "  ${RED}✗${NC} Credenciais AWS não configuradas"
    echo -e "     Execute: aws configure"
    ((errors++))
fi

echo ""
echo "📝 Verificando permissões AWS..."

check_permission() {
    local service=$1
    local action=$2
    
    aws $service $action --max-items 1 >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} $service:$action"
    else
        echo -e "  ${YELLOW}⚠${NC} $service:$action (verifique permissões)"
    fi
}

check_permission "eks" "list-clusters"
check_permission "rds" "describe-db-instances"
check_permission "secretsmanager" "list-secrets"
check_permission "ec2" "describe-vpcs"

echo ""
echo "📂 Verificando arquivos..."

check_file() {
    if [ -f "$1" ]; then
        echo -e "  ${GREEN}✓${NC} $1"
    else
        echo -e "  ${RED}✗${NC} $1 não encontrado"
        ((errors++))
    fi
}

check_file "scripts/bootstrap.sh"
check_file "scripts/secrets-manager.sh"
check_file "scripts/quick-start.sh"
check_file "infrastructure/terraform/main.tf"
check_file "infrastructure/terraform/variables.tf"

echo ""

if [ $errors -eq 0 ]; then
    echo -e "${GREEN}✅ Ambiente OK! Pronto para deploy${NC}"
    echo ""
    echo "Execute:"
    echo "  make aws-deploy-complete"
    echo ""
    echo "Ou:"
    echo "  ./scripts/quick-start.sh"
    exit 0
else
    echo -e "${RED}❌ $errors erro(s) encontrado(s)${NC}"
    echo ""
    echo "Corrija os problemas acima antes de executar o deploy"
    exit 1
fi

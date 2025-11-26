#!/usr/bin/env bash
set -euo pipefail

# ===========================================================================================
# Velure Secrets Manager
# ===========================================================================================
# Gerenciamento de secrets com AWS Secrets Manager
# 
# Uso: ./scripts/secrets-manager.sh [COMMAND]
# 
# Comandos:
#   create          Cria todos os secrets
#   update          Atualiza secrets existentes
#   get <name>      Busca um secret específico
#   list            Lista todos os secrets
#   delete          Deleta todos os secrets (PERIGOSO!)
#   rotate <name>   Rotaciona um secret específico
#   export-env      Exporta secrets como variáveis de ambiente
# ===========================================================================================

PROJECT_NAME="velure"
ENVIRONMENT="prod"
AWS_REGION="us-east-1"
SECRETS_PREFIX="${PROJECT_NAME}/${ENVIRONMENT}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ${NC} $1"; }
log_success() { echo -e "${GREEN}✓${NC} $1"; }
log_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }

generate_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

create_secret() {
    local name=$1
    local value=$2
    local description=$3
    
    aws secretsmanager create-secret \
        --name "${SECRETS_PREFIX}/${name}" \
        --description "${description}" \
        --secret-string "${value}" \
        --region "${AWS_REGION}" 2>/dev/null && \
        log_success "Secret criado: ${name}" || \
        log_warning "Secret já existe: ${name}"
}

update_secret() {
    local name=$1
    local value=$2
    
    aws secretsmanager update-secret \
        --secret-id "${SECRETS_PREFIX}/${name}" \
        --secret-string "${value}" \
        --region "${AWS_REGION}" && \
        log_success "Secret atualizado: ${name}"
}

get_secret() {
    local name=$1
    aws secretsmanager get-secret-value \
        --secret-id "${SECRETS_PREFIX}/${name}" \
        --region "${AWS_REGION}" \
        --query 'SecretString' \
        --output text
}

list_secrets() {
    log_info "Secrets em ${SECRETS_PREFIX}:"
    aws secretsmanager list-secrets \
        --region "${AWS_REGION}" \
        --query "SecretList[?starts_with(Name, '${SECRETS_PREFIX}/')].Name" \
        --output table
}

delete_secret() {
    local name=$1
    aws secretsmanager delete-secret \
        --secret-id "${SECRETS_PREFIX}/${name}" \
        --force-delete-without-recovery \
        --region "${AWS_REGION}" && \
        log_success "Secret deletado: ${name}"
}

rotate_secret() {
    local name=$1
    local new_password=$(generate_password)
    
    log_info "Rotacionando secret: ${name}"
    
    case $name in
        rds-auth)
            update_secret "${name}" "{\"username\":\"postgres\",\"password\":\"${new_password}\",\"dbname\":\"velure_auth\"}"
            ;;
        rds-orders)
            update_secret "${name}" "{\"username\":\"postgres\",\"password\":\"${new_password}\",\"dbname\":\"velure_orders\"}"
            ;;
        rabbitmq)
            update_secret "${name}" "{\"username\":\"admin\",\"password\":\"${new_password}\"}"
            ;;
        jwt)
            local refresh_secret=$(generate_password)
            update_secret "${name}" "{\"secret\":\"${new_password}\",\"refreshSecret\":\"${refresh_secret}\",\"expiresIn\":\"1h\",\"refreshExpiresIn\":\"7d\"}"
            ;;
        session)
            update_secret "${name}" "{\"secret\":\"${new_password}\",\"expiresIn\":\"86400000\"}"
            ;;
        *)
            log_error "Secret desconhecido: ${name}"
            return 1
            ;;
    esac
    
    log_warning "Atenção: Reinicie os pods para aplicar o novo secret!"
}

create_all_secrets() {
    log_info "Criando todos os secrets..."
    
    local rds_auth_password=$(generate_password)
    local rds_orders_password=$(generate_password)
    local rabbitmq_password=$(generate_password)
    local jwt_secret=$(generate_password)
    local jwt_refresh_secret=$(generate_password)
    local session_secret=$(generate_password)
    
    create_secret "rds-auth" \
        "{\"username\":\"postgres\",\"password\":\"${rds_auth_password}\",\"dbname\":\"velure_auth\"}" \
        "RDS Auth Service Database Credentials"
    
    create_secret "rds-orders" \
        "{\"username\":\"postgres\",\"password\":\"${rds_orders_password}\",\"dbname\":\"velure_orders\"}" \
        "RDS Orders Service Database Credentials"
    
    create_secret "rabbitmq" \
        "{\"username\":\"admin\",\"password\":\"${rabbitmq_password}\"}" \
        "RabbitMQ Admin Credentials"
    
    create_secret "jwt" \
        "{\"secret\":\"${jwt_secret}\",\"refreshSecret\":\"${jwt_refresh_secret}\",\"expiresIn\":\"1h\",\"refreshExpiresIn\":\"7d\"}" \
        "JWT Secrets for Auth Service"
    
    create_secret "session" \
        "{\"secret\":\"${session_secret}\",\"expiresIn\":\"86400000\"}" \
        "Session Secret for Auth Service"
    
    log_success "Todos os secrets criados!"
}

delete_all_secrets() {
    log_warning "DELETANDO TODOS OS SECRETS!"
    read -p "Tem certeza? Digite 'DELETE' para confirmar: " confirm
    [ "$confirm" != "DELETE" ] && exit 0
    
    for secret in rds-auth rds-orders rabbitmq jwt session mongodb; do
        delete_secret "${secret}" 2>/dev/null || true
    done
}

export_env() {
    log_info "Exportando secrets como variáveis de ambiente..."
    
    cat << EOF
# Velure Secrets (Auto-generated)
export RDS_AUTH_PASSWORD="\$(echo '$(get_secret rds-auth)' | jq -r .password)"
export RDS_ORDERS_PASSWORD="\$(echo '$(get_secret rds-orders)' | jq -r .password)"
export RABBITMQ_PASSWORD="\$(echo '$(get_secret rabbitmq)' | jq -r .password)"
export JWT_SECRET="\$(echo '$(get_secret jwt)' | jq -r .secret)"
export JWT_REFRESH_SECRET="\$(echo '$(get_secret jwt)' | jq -r .refreshSecret)"
export SESSION_SECRET="\$(echo '$(get_secret session)' | jq -r .secret)"
EOF
}

case "${1:-}" in
    create) create_all_secrets ;;
    update) shift; update_secret "$@" ;;
    get) shift; get_secret "$@" ;;
    list) list_secrets ;;
    delete) delete_all_secrets ;;
    rotate) shift; rotate_secret "$@" ;;
    export-env) export_env ;;
    *)
        echo "Uso: $0 {create|update|get|list|delete|rotate|export-env}"
        exit 1
        ;;
esac

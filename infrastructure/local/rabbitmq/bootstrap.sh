#!/usr/bin/env bash
set -euo pipefail

: "${PUBLISHER_RABBITMQ_USER:?env faltando}"
: "${PUBLISHER_RABBITMQ_PASSWORD:?env faltando}"
: "${PROCESS_RABBITMQ_USER:?env faltando}"
: "${PROCESS_RABBITMQ_PASSWORD:?env faltando}"
: "${ADMIN_RABBITMQ_USER:?env faltando}"
: "${ADMIN_RABBITMQ_PASSWORD:?env faltando}"

echo "[bootstrap] Iniciando rabbitmq-server em background..."
rabbitmq-server &            # inicia o broker
RABBIT_PID=$!

echo "[bootstrap] Aguardando RabbitMQ ficar pronto..."
until rabbitmq-diagnostics -q ping; do sleep 2; done
rabbitmqctl await_startup
echo "[bootstrap] OK."

ensure_user() {
  local user="$1" pass="$2"
  if rabbitmqctl list_users | awk '{print $1}' | grep -xq "$user"; then
    echo "[bootstrap] User '$user' existe — atualizando senha."
    rabbitmqctl change_password "$user" "$pass"
  else
    echo "[bootstrap] Criando user '$user'."
    rabbitmqctl add_user "$user" "$pass"
  fi
}

set_perms() {
  local user="$1" vhost="${2:-/}"
  rabbitmqctl set_permissions -p "$vhost" "$user" ".*" ".*" ".*"
}

# usuários de app
ensure_user "$PUBLISHER_RABBITMQ_USER" "$PUBLISHER_RABBITMQ_PASSWORD"
set_perms   "$PUBLISHER_RABBITMQ_USER" /

ensure_user "$PROCESS_RABBITMQ_USER" "$PROCESS_RABBITMQ_PASSWORD"
set_perms   "$PROCESS_RABBITMQ_USER" /

# admin pro management
ensure_user "$ADMIN_RABBITMQ_USER" "$ADMIN_RABBITMQ_PASSWORD"
rabbitmqctl set_user_tags "$ADMIN_RABBITMQ_USER" administrator
set_perms   "$ADMIN_RABBITMQ_USER" /

# opcional: remover guest
rabbitmqctl delete_user guest || true

echo "[bootstrap] Declarando exchange 'orders'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare exchange name=orders type=topic durable=true

echo "[bootstrap] Declarando fila 'process-order-queue'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare queue name=process-order-queue durable=true

echo "[bootstrap] Criando binding entre exchange 'orders' e fila 'process-order-queue'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare binding source=orders destination=process-order-queue routing_key=order.created

echo "[bootstrap] Infraestrutura RabbitMQ configurada com sucesso!"

echo "[bootstrap] Concluído. Aguardando o RabbitMQ (pid=$RABBIT_PID)."
wait "$RABBIT_PID"

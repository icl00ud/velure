#!/usr/bin/env bash
set -euo pipefail

: "${PUBLISHER_RABBITMQ_USER:?env missing}"
: "${PUBLISHER_RABBITMQ_PASSWORD:?env missing}"
: "${PROCESS_RABBITMQ_USER:?env missing}"
: "${PROCESS_RABBITMQ_PASSWORD:?env missing}"
: "${ADMIN_RABBITMQ_USER:?env missing}"
: "${ADMIN_RABBITMQ_PASSWORD:?env missing}"

echo "[bootstrap] Starting rabbitmq-server in the background..."
rabbitmq-server &            # start the broker
RABBIT_PID=$!

echo "[bootstrap] Waiting for RabbitMQ to be ready..."
until rabbitmq-diagnostics -q ping; do sleep 2; done
rabbitmqctl await_startup
echo "[bootstrap] OK."

ensure_user() {
  local user="$1" pass="$2"
  if rabbitmqctl list_users | awk '{print $1}' | grep -xq "$user"; then
    echo "[bootstrap] User '$user' exists — updating password."
    rabbitmqctl change_password "$user" "$pass"
  else
    echo "[bootstrap] Creating user '$user'."
    rabbitmqctl add_user "$user" "$pass"
  fi
}

set_perms() {
  local user="$1" vhost="${2:-/}"
  rabbitmqctl set_permissions -p "$vhost" "$user" ".*" ".*" ".*"
}

# Application users
ensure_user "$PUBLISHER_RABBITMQ_USER" "$PUBLISHER_RABBITMQ_PASSWORD"
set_perms   "$PUBLISHER_RABBITMQ_USER" /

ensure_user "$PROCESS_RABBITMQ_USER" "$PROCESS_RABBITMQ_PASSWORD"
set_perms   "$PROCESS_RABBITMQ_USER" /

# Admin user for the management UI
ensure_user "$ADMIN_RABBITMQ_USER" "$ADMIN_RABBITMQ_PASSWORD"
rabbitmqctl set_user_tags "$ADMIN_RABBITMQ_USER" administrator
set_perms   "$ADMIN_RABBITMQ_USER" /

# Optional: remove the default guest user
rabbitmqctl delete_user guest || true

echo "[bootstrap] Declaring exchange 'orders'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare exchange name=orders type=topic durable=true

# Dead Letter Exchange (DLX) for permanently failed messages
echo "[bootstrap] Declaring Dead Letter Exchange 'orders.dlx'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare exchange name=orders.dlx type=fanout durable=true

# Dead Letter Queue (DLQ) that stores rejected messages
echo "[bootstrap] Declaring Dead Letter Queue 'process-order-queue.dlq'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare queue name=process-order-queue.dlq durable=true

# Bind the DLQ to the DLX
echo "[bootstrap] Binding DLQ to DLX..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare binding source=orders.dlx destination=process-order-queue.dlq

# Main queue configured with the DLX so rejected messages flow to the DLQ
echo "[bootstrap] Declaring queue 'process-order-queue' with DLX..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare queue name=process-order-queue durable=true \
  arguments='{"x-dead-letter-exchange":"orders.dlx","x-max-length":10000}'

echo "[bootstrap] Creating binding between exchange 'orders' and queue 'process-order-queue'..."
rabbitmqadmin -u "$ADMIN_RABBITMQ_USER" -p "$ADMIN_RABBITMQ_PASSWORD" declare binding source=orders destination=process-order-queue routing_key=order.created

echo "[bootstrap] RabbitMQ infrastructure configured successfully!"

echo "[bootstrap] Done. Waiting on RabbitMQ (pid=$RABBIT_PID)."
wait "$RABBIT_PID"

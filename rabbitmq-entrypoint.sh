#!/bin/bash
set -e

rabbitmq-server -detached

sleep 10

if [ -n "$PUBLISHER_RABBITMQ_PASSWORD" ]; then
  rabbitmqctl add_user "${PUBLISHER_RABBITMQ_USER}" "$PUBLISHER_RABBITMQ_PASSWORD" || true
  rabbitmqctl set_permissions -p "/" "${PUBLISHER_RABBITMQ_USER}" ".*" ".*" ".*"
fi

if [ -n "$PROCESS_RABBITMQ_PASSWORD" ]; then
  rabbitmqctl add_user "${PROCESS_RABBITMQ_USER}" "$PROCESS_RABBITMQ_PASSWORD" || true
  rabbitmqctl set_permissions -p "/" "${PROCESS_RABBITMQ_USER}" ".*" ".*" ".*"
fi

rabbitmqctl stop

exec rabbitmq-server

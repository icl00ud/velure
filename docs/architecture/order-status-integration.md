# Order Status Integration

## Visão Geral

Este documento descreve a integração entre os microserviços `process-order-service` e `publish-order-service` para atualização automática de status de pedidos.

## Arquitetura

### Fluxo de Eventos

1. **publish-order-service** cria um pedido com status `CREATED`
2. **publish-order-service** publica evento `order.created` no RabbitMQ
3. **process-order-service** consome evento `order.created`
4. **process-order-service** processa pagamento e publica:
   - Evento `order.processing` (status intermediário)
   - Evento `order.completed` (após processamento)
5. **publish-order-service** consome eventos e atualiza status no PostgreSQL

### Eventos

#### order.created
```json
{
  "type": "order.created",
  "payload": {
    "id": "uuid",
    "items": [...],
    "total": 1000,
    "status": "CREATED"
  }
}
```

#### order.processing
```json
{
  "type": "order.processing",
  "payload": {
    "id": "uuid"
  }
}
```

#### order.completed
```json
{
  "type": "order.completed",
  "payload": {
    "id": "uuid",
    "order_id": "uuid",
    "amount": 1000,
    "processed_at": "2025-10-06T12:00:00Z"
  }
}
```

## Componentes

### publish-order-service

#### Consumer (novo)
- **Arquivo**: `internal/consumer/rabbitmq_consumer.go`
- **Função**: Consome eventos `order.processing` e `order.completed`
- **Queue**: `publish-order-status-updates` (configurável)
- **Workers**: 3 (configurável via `PUBLISHER_CONSUMER_WORKERS`)

#### Event Handler (novo)
- **Arquivo**: `internal/handler/event_handler.go`
- **Função**: Processa eventos e atualiza status no PostgreSQL

#### Status Constants (novo)
- `StatusCreated`: "CREATED"
- `StatusProcessing`: "PROCESSING"
- `StatusCompleted`: "COMPLETED"

### process-order-service

#### Publisher
- **Arquivo**: `internal/service/payment_service.go`
- **Eventos publicados**:
  - `order.processing`: Quando inicia processamento
  - `order.completed`: Quando finaliza processamento

## Configuração

### Variáveis de Ambiente

#### publish-order-service
```bash
PUBLISHER_ORDER_SERVICE_APP_PORT=3030
PUBLISHER_RABBITMQ_URL=amqp://publisher:password@rabbitmq:5672/
ORDER_EXCHANGE=order_events
PUBLISHER_ORDER_QUEUE=publish-order-status-updates
PUBLISHER_CONSUMER_WORKERS=3
POSTGRES_URL=postgresql://user:pass@postgres:5432/orders
```

#### process-order-service
```bash
PROCESS_ORDER_SERVICE_APP_PORT=3040
PROCESS_RABBITMQ_URL=amqp://processor:password@rabbitmq:5672/
ORDER_EXCHANGE=order_events
RABBITMQ_ORDER_QUEUE=order_created_queue
POSTGRES_URL=postgresql://user:pass@postgres:5432/orders
```

### RabbitMQ

#### Exchange
- **Nome**: `order_events`
- **Tipo**: `topic`
- **Durável**: `true`

#### Queues
- `order_created_queue`: Para process-order-service
- `publish-order-status-updates`: Para publish-order-service

#### Bindings
- `order_created_queue` → `order.created`
- `publish-order-status-updates` → `order.processing`
- `publish-order-status-updates` → `order.completed`

## Observabilidade

### Logs Estruturados

Todos os eventos são logados com contexto:

```json
{
  "level": "info",
  "msg": "event processed successfully",
  "type": "order.processing",
  "order_id": "uuid",
  "worker_id": 1
}
```

### Métricas Recomendadas

- `order_events_consumed_total{type, status}`: Total de eventos consumidos
- `order_status_updates_total{from, to}`: Total de atualizações de status
- `order_processing_duration_seconds`: Duração do processamento

### Health Checks

#### publish-order-service
- `/health`: Health check básico
- `/healthz`: Liveness probe
- `/readyz`: Readiness probe

## Resiliência

### Error Handling
- Eventos com erro são rejeitados com `Nack(requeue=true)`
- Retry automático pelo RabbitMQ
- Dead Letter Queue recomendada para falhas persistentes

### Graceful Shutdown
- Context cancellation para workers
- Finalização ordenada de conexões
- Timeout de 10 segundos

### QoS
- Prefetch count: 1 mensagem por worker
- Garantia de processamento sequencial

## Testing

### Local
```bash
# Build
cd publish-order-service && go build -o bin/publish-order-service .
cd ../process-order-service && go build -o bin/process-order-service .

# Run with Docker Compose
docker-compose up -d rabbitmq postgres
docker-compose up publish-order-service process-order-service
```

### Verificação
```bash
# Criar pedido
curl -X POST http://localhost:3030/create-order \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":"123","name":"Test","quantity":1,"price":100}]}'

# Verificar logs
docker-compose logs -f publish-order-service
docker-compose logs -f process-order-service

# Consultar pedidos
curl http://localhost:3030/orders
```

## Próximos Passos

### Melhorias Recomendadas

1. **Dead Letter Queue**: Implementar DLQ para eventos com falha
2. **Circuit Breaker**: Adicionar circuit breaker para PostgreSQL
3. **Idempotência**: Garantir idempotência no processamento de eventos
4. **Distributed Tracing**: Implementar OpenTelemetry
5. **Schema Validation**: Validar schema dos eventos com JSON Schema
6. **Event Sourcing**: Considerar event sourcing para auditoria completa

### Monitoramento

1. Configurar alertas para:
   - Taxa de erro > 5%
   - Latência > 1s
   - Queue depth > 1000
   - Consumer lag > 100

2. Dashboards Grafana:
   - Order flow overview
   - Event processing metrics
   - Error rates by event type
   - Consumer performance

## Troubleshooting

### Eventos não são processados
- Verificar bindings no RabbitMQ
- Verificar logs do consumer
- Verificar conectividade com PostgreSQL

### Status não atualiza
- Verificar se order_id está correto no payload
- Verificar logs do event handler
- Consultar tabela TBLOrders diretamente

### Performance
- Aumentar número de workers
- Verificar queries PostgreSQL
- Adicionar índices se necessário

## Referências

- [RabbitMQ Topic Exchange](https://www.rabbitmq.com/tutorials/tutorial-five-go.html)
- [Go Context Patterns](https://go.dev/blog/context)
- [PostgreSQL Best Practices](https://wiki.postgresql.org/wiki/Don%27t_Do_This)

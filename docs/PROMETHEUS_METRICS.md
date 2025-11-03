# Prometheus Metrics - Velure Microservices

Este documento descreve todas as métricas Prometheus implementadas nos microserviços do Velure.

## Endpoints de Métricas

Todos os serviços expõem métricas no endpoint `/metrics`:

- **auth-service**: `http://localhost:3020/metrics`
- **product-service**: `http://localhost:3010/metrics`
- **publish-order-service**: `http://localhost:8080/metrics`
- **process-order-service**: `http://localhost:8081/metrics`

## Auth Service

### Métricas Principais

| Métrica | Tipo | Labels | Descrição |
|---------|------|--------|-----------|
| `auth_login_attempts_total` | Counter | status | Total de tentativas de login |
| `auth_login_duration_seconds` | Histogram | status | Duração do login |
| `auth_registration_attempts_total` | Counter | status | Total de registros |
| `auth_token_validations_total` | Counter | result | Total de validações de token |
| `auth_token_generations_total` | Counter | - | Total de tokens gerados |
| `auth_active_sessions` | Gauge | - | Sessões ativas |
| `auth_total_users` | Gauge | - | Total de usuários |
| `auth_errors_total` | Counter | type | Total de erros |
| `auth_http_requests_total` | Counter | method, endpoint, status | Total de requests HTTP |

## Product Service

### Métricas Principais

| Métrica | Tipo | Labels | Descrição |
|---------|------|--------|-----------|
| `product_queries_total` | Counter | operation | Total de consultas de produtos |
| `product_mutations_total` | Counter | operation, status | Total de modificações |
| `product_cache_hits_total` | Counter | - | Cache hits |
| `product_cache_misses_total` | Counter | - | Cache misses |
| `product_inventory_updates_total` | Counter | status | Atualizações de inventário |
| `product_catalog_total` | Gauge | - | Total de produtos no catálogo |
| `product_searches_total` | Counter | type | Total de buscas |
| `product_search_results_count` | Histogram | - | Número de resultados retornados |
| `product_errors_total` | Counter | type | Total de erros |
| `product_http_requests_total` | Counter | method, path, status | Total de requests HTTP |

## Publish Order Service

### Métricas Principais

| Métrica | Tipo | Labels | Descrição |
|---------|------|--------|-----------|
| `publish_order_created_total` | Counter | status | Total de pedidos criados |
| `publish_order_creation_duration_seconds` | Histogram | - | Duração da criação de pedidos |
| `publish_order_published_total` | Counter | status | Total de pedidos publicados |
| `publish_order_rabbitmq_publish_duration_seconds` | Histogram | - | Duração de publicação no RabbitMQ |
| `publish_order_status_updates_total` | Counter | old_status, new_status | Atualizações de status |
| `publish_order_current_by_status` | Gauge | status | Pedidos por status |
| `publish_order_sse_connections` | Gauge | - | Conexões SSE ativas |
| `publish_order_sse_messages_sent_total` | Counter | - | Mensagens SSE enviadas |
| `publish_order_total_value` | Histogram | - | Valor total dos pedidos |
| `publish_order_items_count` | Histogram | - | Número de itens por pedido |
| `publish_order_errors_total` | Counter | type | Total de erros |

## Process Order Service

### Métricas Principais

| Métrica | Tipo | Labels | Descrição |
|---------|------|--------|-----------|
| `process_order_processed_total` | Counter | status | Total de pedidos processados |
| `process_order_processing_duration_seconds` | Histogram | - | Duração do processamento |
| `process_order_payment_attempts_total` | Counter | result | Tentativas de pagamento |
| `process_order_payment_processing_duration_seconds` | Histogram | - | Duração do processamento de pagamento |
| `process_order_payment_value` | Histogram | - | Valor dos pagamentos |
| `process_order_inventory_checks_total` | Counter | result | Checagens de inventário |
| `process_order_inventory_check_duration_seconds` | Histogram | - | Duração da checagem de inventário |
| `process_order_messages_consumed_total` | Counter | - | Mensagens consumidas do RabbitMQ |
| `process_order_message_processing_errors_total` | Counter | - | Erros de processamento |
| `process_order_active_workers` | Gauge | - | Workers ativos |
| `process_order_product_service_calls_total` | Counter | operation, status | Chamadas ao product-service |
| `process_order_errors_total` | Counter | type | Total de erros |

## Queries PromQL Úteis

### Taxa de Sucesso de Login
```promql
rate(auth_login_attempts_total{status="success"}[5m]) /
rate(auth_login_attempts_total[5m])
```

### Latência p95 de Login
```promql
histogram_quantile(0.95, rate(auth_login_duration_seconds_bucket[5m]))
```

### Cache Hit Rate (Product Service)
```promql
rate(product_cache_hits_total[5m]) /
(rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))
```

### Taxa de Sucesso de Pagamentos
```promql
rate(process_order_payment_attempts_total{result="success"}[5m]) /
rate(process_order_payment_attempts_total[5m])
```

### Pedidos Processados por Minuto
```promql
rate(process_order_processed_total[1m]) * 60
```

### Taxa de Erros Geral
```promql
sum(rate(auth_errors_total[5m])) +
sum(rate(product_errors_total[5m])) +
sum(rate(publish_order_errors_total[5m])) +
sum(rate(process_order_errors_total[5m]))
```

### Throughput HTTP por Serviço
```promql
sum by (service) (rate(auth_http_requests_total[1m]))
sum by (service) (rate(product_http_requests_total[1m]))
sum by (service) (rate(publish_order_http_requests_total[1m]))
```

### Latência p99 por Endpoint (Auth)
```promql
histogram_quantile(0.99,
  sum by (endpoint, le) (rate(auth_http_request_duration_seconds_bucket[5m]))
)
```

## Alertas Recomendados

### Alta Taxa de Falhas de Login
```yaml
- alert: HighLoginFailureRate
  expr: |
    rate(auth_login_attempts_total{status="failure"}[5m]) > 10
  for: 2m
  annotations:
    summary: "Alta taxa de falhas de login"
```

### Baixo Cache Hit Rate
```yaml
- alert: LowCacheHitRate
  expr: |
    rate(product_cache_hits_total[5m]) /
    (rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m])) < 0.5
  for: 5m
  annotations:
    summary: "Taxa de cache hit abaixo de 50%"
```

### Alta Taxa de Erros
```yaml
- alert: HighErrorRate
  expr: |
    sum(rate(process_order_errors_total[5m])) > 5
  for: 2m
  annotations:
    summary: "Alta taxa de erros no processamento de pedidos"
```

### Fila de Pedidos Crescendo
```yaml
- alert: OrderQueueGrowing
  expr: |
    process_order_queue_size > 100
  for: 5m
  annotations:
    summary: "Fila de pedidos crescendo além do normal"
```

### Pagamentos Falhando
```yaml
- alert: PaymentFailureRate
  expr: |
    rate(process_order_payment_attempts_total{result="failure"}[5m]) /
    rate(process_order_payment_attempts_total[5m]) > 0.1
  for: 3m
  annotations:
    summary: "Mais de 10% dos pagamentos falhando"
```

## Integração com Grafana

Veja `docs/MONITORING.md` para instruções sobre como visualizar essas métricas no Grafana.

Os dashboards pré-configurados estão disponíveis em:
- `infrastructure/kubernetes/monitoring/dashboards/auth-service.json`
- `infrastructure/kubernetes/monitoring/dashboards/product-service.json`
- `infrastructure/kubernetes/monitoring/dashboards/orders-service.json`
- `infrastructure/kubernetes/monitoring/dashboards/overview.json`

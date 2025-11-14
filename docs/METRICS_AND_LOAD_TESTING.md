# Métricas Implementadas e Teste de Carga

## ✅ Métricas Implementadas

### Auth Service
**RED Metrics (Rate, Errors, Duration):**
- `auth_http_requests_total` - Total de requisições HTTP (method, path, status)
- `auth_http_request_duration_seconds` - Duração de requisições HTTP
- `auth_errors_total` - Total de erros por tipo

**Business Metrics:**
- `auth_login_attempts_total` - Tentativas de login (success/failure)
- `auth_login_duration_seconds` - Duração de logins
- `auth_registration_attempts_total` - Tentativas de registro
- `auth_registration_duration_seconds` - Duração de registros
- `auth_token_generations_total` - Total de tokens gerados
- `auth_token_generation_duration_seconds` - Duração de geração de tokens
- `auth_logout_requests_total` - Total de logouts
- `auth_total_users` - Total de usuários registrados (gauge)

### Product Service
**RED Metrics:**
- `product_http_requests_total` - Total de requisições HTTP
- `product_http_request_duration_seconds` - Duração de requisições HTTP
- `product_errors_total` - Total de erros

**Business Metrics:**
- `product_queries_total` - Total de queries por operação
- `product_mutations_total` - Total de mutations (create, update, delete)
- `product_operation_duration_seconds` - Duração de operações
- `product_cache_hits_total` - Cache hits
- `product_cache_misses_total` - Cache misses
- `product_catalog_total` - Total de produtos no catálogo (gauge)
- `product_searches_total` - Total de buscas
- `product_category_queries_total` - Queries por categoria

### Publish Order Service
**RED Metrics:**
- `publish_order_http_requests_total` - Total de requisições HTTP
- `publish_order_http_request_duration_seconds` - Duração de requisições
- `publish_order_errors_total` - Total de erros

**Business Metrics:**
- `publish_order_created_total` - Ordens criadas (success/failure)
- `publish_order_creation_duration_seconds` - Duração de criação
- `publish_order_rabbitmq_publish_duration_seconds` - Duração de publicação no RabbitMQ
- `publish_order_sse_messages_sent_total` - Mensagens SSE enviadas
- `publish_order_total_value` - Valor total das ordens (histogram)

### Process Order Service
**Metrics:**
- `process_order_messages_consumed_total` - Mensagens consumidas do RabbitMQ
- `process_order_message_processing_errors_total` - Erros de processamento
- `process_order_processing_duration_seconds` - Duração de processamento
- `process_order_payment_processing_duration_seconds` - Duração de processamento de pagamento
- `process_order_inventory_check_duration_seconds` - Duração de verificação de estoque

## 🧪 Testes de Carga Disponíveis

### 1. Teste Simplificado (`simple-load-test.js`)
**Objetivo:** Validação básica de saúde e funcionalidade dos serviços

**Configuração:**
- VUs: 10
- Duração: 2m
- Thresholds:
  - `http_req_duration p(95) < 500ms`
  - `http_req_failed rate < 1%`
  - `auth_success_rate > 90%`
  - `product_success_rate > 95%`

**Como executar:**
```bash
k6 run --duration 1m --vus 10 tests/load/simple-load-test.js
```

**Resultados recentes:**
- ✅ 616 checks totais (100% sucesso)
- ✅ P95 latência: 1.33ms
- ✅ P99 latência: 5.25ms
- ✅ Taxa de erro: 0%

### 2. Teste Integrado Completo (`full-system-load-test.js`)
**Objetivo:** Teste end-to-end com múltiplos cenários

**Cenários:**
1. **Browse Products (30%)** - Navegação básica de produtos
2. **Register and Browse (30%)** - Registro + navegação autenticada
3. **Full Purchase (40%)** - Fluxo completo de compra

**Configuração avançada:**
- Stages graduais: 10 → 50 → 100 VUs
- Duração total: 17 minutos
- Thresholds:
  - `http_req_duration p(95) < 500ms`
  - `http_req_duration p(99) < 1000ms`
  - `http_req_failed rate < 5%`
  - `order_creation_success_rate > 90%`

**Como executar:**
```bash
k6 run --duration 3m --vus 30 tests/load/full-system-load-test.js
```

## 📊 Dashboards e Visualização

### Prometheus Queries Úteis

**Taxa de requisições por serviço:**
```promql
rate(auth_http_requests_total[5m])
rate(product_http_requests_total[5m])
```

**Latência P95 por serviço:**
```promql
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(product_http_request_duration_seconds_bucket[5m]))
```

**Taxa de erro:**
```promql
rate(auth_errors_total[5m])
rate(product_errors_total[5m])
```

**Cache hit rate (Product Service):**
```promql
rate(product_cache_hits_total[5m]) / 
  (rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))
```

### Acessar Grafana
```bash
open http://localhost:3000
```

**Credentials:**
- User: `admin`
- Password: `admin`

## 🎯 Métricas SLO/SLI Recomendadas

### Availability
- **Target:** 99.9% uptime
- **Metric:** `1 - (rate(http_req_failed[5m]) / rate(http_reqs[5m]))`

### Latency
- **Target:** P95 < 500ms, P99 < 1000ms
- **Metric:** `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`

### Error Rate
- **Target:** < 1%
- **Metric:** `rate(errors_total[5m]) / rate(http_reqs[5m]) * 100`

### Throughput
- **Target:** > 100 req/s por serviço
- **Metric:** `rate(http_requests_total[1m])`

## 🔍 Comandos Úteis

### Verificar métricas disponíveis:
```bash
curl -s 'http://localhost:9090/api/v1/label/__name__/values' | jq -r '.data[]' | grep -E "^(auth|product|publish|process)_" | sort
```

### Query específica de métrica:
```bash
curl -s 'http://localhost:9090/api/v1/query?query=auth_http_requests_total' | jq '.'
```

### Monitorar logs em tempo real:
```bash
docker-compose -f infrastructure/local/docker-compose.yaml logs -f auth-service product-service
```

## 📝 Notas Importantes

1. **Endpoints Corretos:**
   - Auth Service: `/authentication/register`, `/authentication/login`
   - Product Service: `/products`
   - Publish Order Service: `/orders`

2. **Health Checks:**
   - Todos os serviços expõem `/health` (não `/healthz`)

3. **Métricas Prometheus:**
   - Expostas em `/metrics` em cada serviço
   - Scraped automaticamente pelo Prometheus local

4. **Formato de Métricas:**
   - Seguem convenção OpenMetrics
   - Labels padronizados: `method`, `path`, `status`, `operation`
   - Nomenclatura snake_case com sufixos semânticos (`_total`, `_seconds`, `_bytes`)

## 🚀 Próximos Passos

1. **Alertas:**
   - Configurar alertmanager para SLO violations
   - Notificações via Slack/Email

2. **Tracing:**
   - Implementar OpenTelemetry para distributed tracing
   - Correlacionar traces com métricas

3. **Dashboards:**
   - Criar dashboards Grafana específicos por serviço
   - Dashboard agregado com visão geral do sistema

4. **Load Testing CI/CD:**
   - Integrar testes de carga no pipeline
   - Performance regression tests

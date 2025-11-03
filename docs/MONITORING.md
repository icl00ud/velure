# Guia de Monitoramento - Velure

Este guia explica como acessar e usar as ferramentas de monitoramento do Velure.

## 📊 Stack de Monitoramento

O Velure usa a stack kube-prometheus, que inclui:

- **Prometheus**: Coleta e armazena métricas
- **Grafana**: Visualização de métricas em dashboards
- **Alertmanager**: Gerenciamento e roteamento de alertas
- **Node Exporter**: Métricas dos nodes do Kubernetes
- **Kube State Metrics**: Métricas dos recursos do Kubernetes

## 🚀 Acesso Rápido

### Prometheus

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
```
**URL**: http://localhost:9090

**O que fazer no Prometheus:**
- Ver targets sendo monitorados: http://localhost:9090/targets
- Executar queries PromQL: http://localhost:9090/graph
- Ver alertas ativos: http://localhost:9090/alerts
- Ver configuração: http://localhost:9090/config

### Grafana

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
```
**URL**: http://localhost:3000
**Credenciais**: admin / admin

**Dashboards disponíveis:**
- **Velure → Overview**: Visão geral de todos os serviços
- **Kubernetes → Compute Resources**: Métricas do cluster
- **Node Exporter**: Métricas dos nodes

### Alertmanager

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-alertmanager 9093:9093
```
**URL**: http://localhost:9093

**O que fazer no Alertmanager:**
- Ver alertas ativos
- Silenciar alertas temporariamente
- Ver histórico de notificações

## 📈 Métricas Disponíveis

### Auth Service

| Métrica | Tipo | Descrição |
|---------|------|-----------|
| `auth_login_attempts_total` | Counter | Tentativas de login |
| `auth_login_duration_seconds` | Histogram | Duração do login |
| `auth_token_validations_total` | Counter | Validações de token |
| `auth_active_sessions` | Gauge | Sessões ativas |
| `auth_total_users` | Gauge | Total de usuários |
| `auth_errors_total` | Counter | Total de erros |

### Product Service

| Métrica | Tipo | Descrição |
|---------|------|-----------|
| `product_queries_total` | Counter | Consultas de produtos |
| `product_cache_hits_total` | Counter | Cache hits |
| `product_cache_misses_total` | Counter | Cache misses |
| `product_inventory_updates_total` | Counter | Atualizações de inventário |
| `product_catalog_total` | Gauge | Total de produtos |
| `product_searches_total` | Counter | Buscas realizadas |

### Order Services

| Métrica | Tipo | Descrição |
|---------|------|-----------|
| `publish_order_created_total` | Counter | Pedidos criados |
| `publish_order_published_total` | Counter | Pedidos publicados no RabbitMQ |
| `process_order_processed_total` | Counter | Pedidos processados |
| `process_order_payment_attempts_total` | Counter | Tentativas de pagamento |
| `process_order_messages_consumed_total` | Counter | Mensagens consumidas |

**Lista completa**: Ver [PROMETHEUS_METRICS.md](./PROMETHEUS_METRICS.md)

## 🔍 Queries PromQL Úteis

### Saúde dos Serviços

```promql
# Serviços UP/DOWN
up{job=~"velure-.*"}

# Uptime
(time() - process_start_time_seconds{job=~"velure-.*"})
```

### Performance

```promql
# Requests por segundo
rate(auth_http_requests_total[5m])

# Latência p95
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))

# Latência p99
histogram_quantile(0.99, rate(auth_http_request_duration_seconds_bucket[5m]))

# Média de latência
rate(auth_http_request_duration_seconds_sum[5m]) /
rate(auth_http_request_duration_seconds_count[5m])
```

### Erros

```promql
# Taxa de erros
rate(auth_errors_total[5m])

# Taxa de erros HTTP 5xx
rate(auth_http_requests_total{status=~"5.."}[5m])

# Percentual de erros
(rate(auth_http_requests_total{status=~"5.."}[5m]) /
 rate(auth_http_requests_total[5m])) * 100
```

### Cache (Product Service)

```promql
# Cache hit rate
rate(product_cache_hits_total[5m]) /
(rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))

# Cache miss rate
rate(product_cache_misses_total[5m]) /
(rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))

# Total de operações de cache
rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m])
```

### Autenticação

```promql
# Taxa de sucesso de login
rate(auth_login_attempts_total{status="success"}[5m]) /
rate(auth_login_attempts_total[5m])

# Login attempts por minuto
rate(auth_login_attempts_total[1m]) * 60

# Sessões ativas
auth_active_sessions

# Crescimento de usuários
increase(auth_total_users[1d])
```

### Pedidos

```promql
# Pedidos por minuto
rate(publish_order_created_total[1m]) * 60

# Taxa de sucesso de pagamentos
rate(process_order_payment_attempts_total{result="success"}[5m]) /
rate(process_order_payment_attempts_total[5m])

# Tempo médio de processamento
rate(process_order_processing_duration_seconds_sum[5m]) /
rate(process_order_processing_duration_seconds_count[5m])

# Tamanho da fila
process_order_queue_size
```

### Recursos (Kubernetes)

```promql
# CPU usage por pod
rate(container_cpu_usage_seconds_total{namespace="default"}[5m])

# Memória usage por pod
container_memory_working_set_bytes{namespace="default"} / 1024 / 1024

# Network I/O
rate(container_network_receive_bytes_total{namespace="default"}[5m])
rate(container_network_transmit_bytes_total{namespace="default"}[5m])
```

## 🎨 Criando Dashboards Customizados

### 1. Acessar Grafana

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
```

### 2. Criar Novo Dashboard

1. Clique em **+** → **Dashboard**
2. Clique em **Add new panel**
3. Selecione datasource **Prometheus**
4. Digite a query PromQL
5. Configure visualização (Graph, Gauge, Stat, etc.)
6. Clique em **Apply**

### 3. Exemplo: Painel de Login Rate

**Query**:
```promql
sum(rate(auth_login_attempts_total[5m])) by (status)
```

**Configurações**:
- Visualization: Time series
- Legend: {{status}}
- Y-axis: requests/sec

### 4. Importar Dashboard Pronto

1. Vá em **Dashboards** → **Import**
2. Upload do arquivo JSON: `infrastructure/kubernetes/monitoring/dashboards/overview-dashboard.json`
3. Selecione Prometheus datasource
4. Clique em **Import**

## 🚨 Alertas Configurados

### Alta Taxa de Falhas de Login
```yaml
Expressão: rate(auth_login_attempts_total{status="failure"}[5m]) > 10
Duração: 2 minutos
Severidade: warning
```

### Baixo Cache Hit Rate
```yaml
Expressão: cache_hit_rate < 0.5
Duração: 5 minutos
Severidade: warning
```

### Serviço Down
```yaml
Expressão: up{job="velure-auth"} == 0
Duração: 1 minuto
Severidade: critical
```

### Alta Taxa de Erros de Pagamento
```yaml
Expressão: payment_failure_rate > 0.1
Duração: 3 minutos
Severidade: warning
```

**Ver todos os alertas**: http://localhost:9090/alerts (Prometheus)

## 🔔 Configurar Notificações

### Slack

Edite o ConfigMap do Alertmanager:

```bash
kubectl edit configmap alertmanager-kube-prometheus-stack-alertmanager -n monitoring
```

Adicione:

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
        title: 'Alert: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

### Email

```yaml
receivers:
  - name: 'email'
    email_configs:
      - to: 'team@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'your@gmail.com'
        auth_password: 'app-password'
```

## 📱 Acesso Externo (Opcional)

### Opção 1: LoadBalancer

Edite o service do Grafana:

```bash
kubectl edit svc kube-prometheus-stack-grafana -n monitoring
```

Mude `type: ClusterIP` para `type: LoadBalancer`

### Opção 2: Ingress

Crie um Ingress para o Grafana:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grafana-ingress
  namespace: monitoring
  annotations:
    kubernetes.io/ingress.class: alb
spec:
  rules:
    - host: grafana.velure.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kube-prometheus-stack-grafana
                port:
                  number: 80
```

## 🔒 Segurança

### Mudar Senha do Grafana

```bash
# Via kubectl
kubectl exec -it -n monitoring <grafana-pod> -- grafana-cli admin reset-admin-password newpassword

# Ou edite o secret
kubectl edit secret kube-prometheus-stack-grafana -n monitoring
```

### Autenticação OAuth (GitHub, Google)

Edite o ConfigMap do Grafana:

```bash
kubectl edit configmap kube-prometheus-stack-grafana -n monitoring
```

Adicione configuração OAuth em `grafana.ini`.

## 📊 Retenção de Dados

### Prometheus

**Padrão**: 15 dias

**Alterar**:

```bash
kubectl edit prometheus kube-prometheus-stack-prometheus -n monitoring
```

```yaml
spec:
  retention: 30d
  retentionSize: 50GB
```

### Grafana

Dashboards e configurações são persistidos em PVC.

**Backup**:

```bash
kubectl exec -n monitoring <grafana-pod> -- \
  tar czf /tmp/grafana-backup.tar.gz /var/lib/grafana
kubectl cp monitoring/<grafana-pod>:/tmp/grafana-backup.tar.gz ./grafana-backup.tar.gz
```

## 🧹 Manutenção

### Limpar Métricas Antigas

```bash
# Prometheus faz isso automaticamente baseado em retention
# Para forçar limpeza:
kubectl exec -n monitoring prometheus-kube-prometheus-stack-prometheus-0 -- \
  promtool tsdb clean-tombstones /prometheus
```

### Verificar Uso de Disco

```bash
kubectl exec -n monitoring prometheus-kube-prometheus-stack-prometheus-0 -- \
  du -sh /prometheus
```

## 📚 Referências

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Tutorial](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Velure Metrics Reference](./PROMETHEUS_METRICS.md)
- [Deploy Guide](./DEPLOY_GUIDE.md)

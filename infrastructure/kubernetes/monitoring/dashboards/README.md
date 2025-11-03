# Grafana Dashboards - Velure

Este diretório contém dashboards pré-configurados para monitoramento dos microserviços Velure.

## Dashboards Disponíveis

### 1. Overview Dashboard (overview-dashboard.json)
Visão geral de todos os serviços com:
- Status de cada serviço (UP/DOWN)
- Taxa de requisições HTTP por serviço
- Latência p95 de resposta
- Taxa de erros por tipo
- Cache hit rate (Product Service)
- Login success rate (Auth Service)

**Quando usar**: Visão rápida da saúde geral do sistema

## Como Importar Dashboards

### Método 1: Via UI do Grafana (Recomendado)

1. Acesse o Grafana:
   ```bash
   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
   ```
   Abra: http://localhost:3000 (admin/admin)

2. No menu lateral, clique em **Dashboards** → **Import**

3. Clique em **Upload JSON file**

4. Selecione o arquivo `overview-dashboard.json`

5. Selecione o datasource **Prometheus**

6. Clique em **Import**

### Método 2: Via ConfigMap (Automático)

```bash
# Criar ConfigMap com todos os dashboards
kubectl create configmap velure-grafana-dashboards \
  --from-file=./overview-dashboard.json \
  -n monitoring

# Adicionar label para o Grafana detectar
kubectl label configmap velure-grafana-dashboards \
  grafana_dashboard=1 \
  -n monitoring

# Reiniciar Grafana para carregar
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n monitoring
```

Os dashboards aparecerão automaticamente na pasta **Velure** do Grafana.

## Criar Dashboards Customizados

### Usando Queries PromQL

Você pode criar seus próprios dashboards usando as queries da documentação:
- Ver: `docs/PROMETHEUS_METRICS.md`

### Exemplos de Queries Úteis

**Taxa de Requisições:**
```promql
sum(rate(auth_http_requests_total[5m])) by (status)
```

**Latência p95:**
```promql
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))
```

**Taxa de Erros:**
```promql
rate(auth_errors_total[5m])
```

**Cache Hit Rate:**
```promql
rate(product_cache_hits_total[5m]) /
(rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))
```

**Pedidos Processados:**
```promql
rate(process_order_processed_total[5m]) * 60
```

## Dashboards Adicionais Recomendados

### Auth Service Dashboard
Crie um dashboard com:
- Login attempts (success vs failure)
- Token validations
- Active sessions
- User registration rate
- Authentication errors

### Product Service Dashboard
Crie um dashboard com:
- Product queries por operação
- Cache performance (hits, misses, operations)
- Inventory updates
- Search performance
- MongoDB query duration

### Orders Dashboard
Crie um dashboard com:
- Orders created vs processed
- Payment success rate
- Payment processing duration
- Order status distribution
- RabbitMQ queue size
- Inventory check failures

### Infrastructure Dashboard
Use dashboards padrão do kube-prometheus-stack:
- **Kubernetes / Compute Resources / Cluster**: Visão geral do cluster
- **Kubernetes / Compute Resources / Namespace (Pods)**: Recursos por namespace
- **Kubernetes / Compute Resources / Pod**: Métricas de pods individuais
- **Node Exporter / Nodes**: Métricas de nodes

## Estrutura de Dashboard Recomendada

```
┌─────────────────────────────────────┐
│  Service Status (UP/DOWN)           │
├─────────────────────────────────────┤
│  Key Metrics (gauges/stats)         │
├─────────────────────────────────────┤
│  Request Rate (timeseries)          │
├─────────────────────────────────────┤
│  Latency p95/p99 (timeseries)       │
├─────────────────────────────────────┤
│  Error Rate (timeseries)            │
├─────────────────────────────────────┤
│  Business Metrics (service-specific)│
└─────────────────────────────────────┘
```

## Troubleshooting

### Dashboards não aparecem

```bash
# Verificar se ConfigMap foi criado
kubectl get configmap velure-grafana-dashboards -n monitoring

# Verificar label
kubectl get configmap velure-grafana-dashboards -n monitoring --show-labels

# Ver logs do Grafana
kubectl logs -n monitoring -l app.kubernetes.io/name=grafana
```

### Métricas não aparecem

```bash
# Verificar se Prometheus está scrapando
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# Acesse: http://localhost:9090/targets

# Verificar ServiceMonitors
kubectl get servicemonitor -A

# Testar endpoint de métricas diretamente
kubectl port-forward <pod-name> 8080:8080
curl http://localhost:8080/metrics
```

### Datasource não conecta

```bash
# Verificar se Prometheus está rodando
kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus

# Testar conectividade
kubectl exec -it -n monitoring <grafana-pod> -- \
  wget -O- http://kube-prometheus-stack-prometheus:9090/-/healthy
```

## Alertas no Grafana

Você pode configurar alertas diretamente nos dashboards:

1. Edite um painel
2. Vá para a aba **Alert**
3. Configure condições (ex: error rate > 5)
4. Configure notification channel (Slack, email, etc.)

**Nota**: Recomendamos usar o Alertmanager do Prometheus para alertas mais robustos (já configurado).

## Referências

- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [Prometheus Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)
- [Velure Prometheus Metrics](../../../../docs/PROMETHEUS_METRICS.md)

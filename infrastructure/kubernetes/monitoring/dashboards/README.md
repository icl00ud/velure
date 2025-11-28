# Grafana Dashboards - Velure

Este diretório contém dashboards pré-configurados para monitoramento dos microserviços Velure.

## Dashboards Disponíveis

- `microservices-overview-dashboard.json`: status dos serviços, taxa de requisições, latência p95/p99, erro %, cache hit rate (Auth e Product), CPU/mem por serviço e profundidade de fila RabbitMQ.
- `overview-dashboard.json`: visão enxuta/legada de saúde geral (status, taxas de erro e latência).
- `auth-service-dashboard.json`: login/registro/token, tráfego HTTP, erros, sessões, cache hit rate (Redis) e hits vs misses.
- `product-service-dashboard.json`: operações de produto, cache (hit/miss + latência), inventário, buscas e MongoDB.
- `publish-order-service-dashboard.json`: criação/publicação de pedidos, SSE, estados, banco de dados e métricas de fila.
- `process-order-service-dashboard.json`: consumo de mensagens, workers, estoque/pagamento, chamadas para Product Service e tamanho da fila.

## Como Importar Dashboards

### Método 1: Via UI do Grafana (Recomendado)

1. Acesse o Grafana:
   ```bash
   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
   ```
   Abra: http://localhost:3000 (admin/admin)

2. No menu lateral, clique em **Dashboards** → **Import**

3. Clique em **Upload JSON file**

4. Selecione o arquivo `.json` desejado (ex: `microservices-overview-dashboard.json`)

5. Selecione o datasource **Prometheus**

6. Clique em **Import**

### Método 2: Via ConfigMap (Automático)

```bash
# Criar/atualizar ConfigMap com TODOS os dashboards deste diretório
kubectl create configmap velure-grafana-dashboards \
  --from-file=infrastructure/kubernetes/monitoring/dashboards/ \
  -n monitoring \
  -o yaml --dry-run=client | kubectl apply -f -

# Adicionar/atualizar label para o Grafana detectar
kubectl label configmap velure-grafana-dashboards \
  grafana_dashboard=1 \
  -n monitoring --overwrite

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

## Dashboards Adicionais / Customizações

- Os dashboards de Auth, Product, Publish Order e Process Order já estão versionados aqui. Duplique-os no Grafana se precisar de filtros/labels específicos do seu ambiente.
- Para infraestrutura use os dashboards padrão do kube-prometheus-stack:
  - **Kubernetes / Compute Resources / Cluster**
  - **Kubernetes / Compute Resources / Namespace (Pods)**
  - **Kubernetes / Compute Resources / Pod**
  - **Node Exporter / Nodes**

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

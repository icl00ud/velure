# Grafana Dashboards - Velure

This directory contains preconfigured dashboards for monitoring the Velure microservices.

## Available Dashboards

- `microservices-overview-dashboard.json`: service status, request rate, p95/p99 latency, error %, cache hit rate (Auth and Product), CPU/memory per service, and RabbitMQ queue depth.
- `overview-dashboard.json`: lean/legacy view of overall health (status, error rate, latency).
- `auth-service-dashboard.json`: login/register/token, HTTP traffic, errors, sessions, cache hit rate (Redis), and hits vs misses.
- `product-service-dashboard.json`: product operations, cache (hit/miss + latency), inventory, searches, and MongoDB.
- `publish-order-service-dashboard.json`: order creation/publishing, SSE, states, database, and queue metrics.
- `process-order-service-dashboard.json`: message consumption, workers, inventory/payment, calls into Product Service, and queue size.

## How to Import Dashboards

### Method 1: via the Grafana UI (recommended)

1. Open Grafana:
   ```bash
   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
   ```
   Then visit: http://localhost:3000 (admin/admin)

2. From the sidebar, click **Dashboards** → **Import**

3. Click **Upload JSON file**

4. Select the desired `.json` file (e.g. `microservices-overview-dashboard.json`)

5. Pick the **Prometheus** datasource

6. Click **Import**

### Method 2: via ConfigMap (automatic)

```bash
# Create/update a ConfigMap with ALL the dashboards in this directory
kubectl create configmap velure-grafana-dashboards \
  --from-file=infrastructure/kubernetes/monitoring/dashboards/ \
  -n monitoring \
  -o yaml --dry-run=client | kubectl apply -f -

# Apply the label Grafana watches for
kubectl label configmap velure-grafana-dashboards \
  grafana_dashboard=1 \
  -n monitoring --overwrite

# Restart Grafana so the dashboards are loaded
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n monitoring
```

The dashboards will appear under the **Velure** folder in Grafana.

## Creating Custom Dashboards

### Using PromQL queries

You can build your own dashboards from the queries documented in:
- `docs/PROMETHEUS_METRICS.md`

### Useful Query Examples

**Request Rate:**
```promql
sum(rate(auth_http_requests_total[5m])) by (status)
```

**p95 Latency:**
```promql
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))
```

**Error Rate:**
```promql
rate(auth_errors_total[5m])
```

**Cache Hit Rate:**
```promql
rate(product_cache_hits_total[5m]) /
(rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))
```

**Orders Processed:**
```promql
rate(process_order_processed_total[5m]) * 60
```

## Additional Dashboards / Customization

- The Auth, Product, Publish Order, and Process Order dashboards are versioned here. Duplicate them in Grafana if you need environment-specific filters/labels.
- For infrastructure, use the dashboards bundled with kube-prometheus-stack:
  - **Kubernetes / Compute Resources / Cluster**
  - **Kubernetes / Compute Resources / Namespace (Pods)**
  - **Kubernetes / Compute Resources / Pod**
  - **Node Exporter / Nodes**

## Recommended Dashboard Structure

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

### Dashboards do not show up

```bash
# Verify the ConfigMap exists
kubectl get configmap velure-grafana-dashboards -n monitoring

# Check the label
kubectl get configmap velure-grafana-dashboards -n monitoring --show-labels

# Check Grafana logs
kubectl logs -n monitoring -l app.kubernetes.io/name=grafana
```

### Metrics do not show up

```bash
# Verify Prometheus is scraping
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# Open: http://localhost:9090/targets

# Check ServiceMonitors
kubectl get servicemonitor -A

# Hit the metrics endpoint directly
kubectl port-forward <pod-name> 8080:8080
curl http://localhost:8080/metrics
```

### Datasource will not connect

```bash
# Check Prometheus is running
kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus

# Test connectivity
kubectl exec -it -n monitoring <grafana-pod> -- \
  wget -O- http://kube-prometheus-stack-prometheus:9090/-/healthy
```

## Grafana Alerts

You can configure alerts directly inside a dashboard:

1. Edit a panel
2. Open the **Alert** tab
3. Configure the condition (e.g. error rate > 5)
4. Configure the notification channel (Slack, email, etc.)

**Note**: prefer Prometheus Alertmanager for robust alerting (already configured).

## References

- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [Prometheus Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)
- [Velure Prometheus Metrics](../../../../docs/PROMETHEUS_METRICS.md)

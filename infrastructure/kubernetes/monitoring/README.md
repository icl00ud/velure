# Velure Monitoring Stack

Complete observability stack for Velure microservices platform using Prometheus, Grafana, and Loki.

## Table of Contents

- [Overview](#overview)
- [Components](#components)
- [Installation](#installation)
  - [Local Development (Docker Compose)](#local-development-docker-compose)
  - [Kubernetes (AWS EKS)](#kubernetes-aws-eks)
- [Dashboards](#dashboards)
- [Alerts](#alerts)
- [Accessing Services](#accessing-services)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Velure monitoring stack provides comprehensive observability across:

- **Metrics**: Prometheus for time-series metrics collection
- **Logs**: Loki for centralized log aggregation
- **Visualization**: Grafana for dashboards and analytics
- **Alerting**: AlertManager for alert routing and notifications

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Grafana (Visualization)                  │
│  - Dashboards        - Alerts         - Explore              │
└────────────┬─────────────────────┬──────────────────────────┘
             │                     │
    ┌────────┴────────┐   ┌────────┴─────────┐
    │   Prometheus    │   │      Loki        │
    │   (Metrics)     │   │     (Logs)       │
    └────────┬────────┘   └────────┬─────────┘
             │                     │
    ┌────────┴────────┐   ┌────────┴─────────┐
    │   Exporters     │   │    Promtail      │
    │  - Node         │   │  (Log collector) │
    │  - cAdvisor     │   └──────────────────┘
    │  - PostgreSQL   │
    │  - MongoDB      │
    │  - Redis        │
    │  - RabbitMQ     │
    └─────────────────┘
             │
    ┌────────┴────────┐
    │ Velure Services │
    │ - Auth          │
    │ - Product       │
    │ - Orders        │
    └─────────────────┘
```

---

## Components

### Core Stack

| Component | Port | Purpose |
|-----------|------|---------|
| **Prometheus** | 9090 | Metrics collection and storage |
| **Grafana** | 3000 | Visualization and dashboards |
| **Loki** | 3100 | Log aggregation system |
| **Promtail** | 9080 | Log collector (ships logs to Loki) |
| **AlertManager** | 9093 | Alert routing and notifications |

### Exporters

| Exporter | Port | Metrics Collected |
|----------|------|-------------------|
| **Node Exporter** | 9100 | Host-level metrics (CPU, memory, disk, network) |
| **cAdvisor** | 8080 | Container metrics |
| **PostgreSQL Exporter** | 9187 | Database connections, queries, cache hits |
| **MongoDB Exporter** | 9216 | Operations, connections, replication |
| **Redis Exporter** | 9121 | Cache hit rate, memory, commands |
| **RabbitMQ Exporter** | 15692 | Queue depth, message rates, connections |

### Service Metrics

All Velure services expose `/metrics` endpoints:

- **auth-service** (3020): Login attempts, token validations, active sessions
- **product-service** (3010): Queries, cache efficiency, catalog size
- **publish-order-service** (8080): Order publications, SSE connections
- **process-order-service** (8081): Order processing, payment simulations

---

## Installation

### Local Development (Docker Compose)

#### Prerequisites

- Docker and Docker Compose installed
- At least 4GB free RAM
- Ports 3000, 3100, 9090 available

#### Quick Start

```bash
# Navigate to infrastructure directory
cd infrastructure/local

# Start infrastructure first
make dev

# Start monitoring stack
docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml up -d

# Verify services are running
docker-compose -f docker-compose.monitoring.yaml ps

# View logs
docker-compose -f docker-compose.monitoring.yaml logs -f
```

#### Access URLs

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Loki**: http://localhost:3100

#### Stop Monitoring

```bash
docker-compose -f docker-compose.monitoring.yaml down

# To remove volumes (delete all metrics/logs)
docker-compose -f docker-compose.monitoring.yaml down -v
```

---

### Kubernetes (AWS EKS)

#### Prerequisites

- Kubernetes cluster running (AWS EKS)
- kubectl configured
- Helm 3 installed
- `monitoring` namespace created

#### Installation Steps

**1. Create monitoring namespace**

```bash
kubectl create namespace monitoring
```

**2. Install kube-prometheus-stack**

```bash
# Add Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install stack
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  -f kube-prometheus-stack-values.yaml \
  -n monitoring
```

**3. Install Loki stack**

```bash
# Add Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install Loki
helm install loki grafana/loki-stack \
  -f loki-stack-values.yaml \
  -n monitoring
```

**4. Deploy database exporters**

```bash
kubectl apply -f database-exporters-servicemonitors.yaml
```

**5. Deploy alert rules**

```bash
kubectl apply -f alert-rules.yaml
kubectl apply -f recording-rules.yaml
```

**6. Configure ServiceMonitors**

```bash
kubectl apply -f auth-service-monitor.yaml
kubectl apply -f product-service-monitor.yaml
kubectl apply -f publish-order-service-monitor.yaml
kubectl apply -f process-order-service-monitor.yaml
```

#### Verification

```bash
# Check Prometheus pods
kubectl get pods -n monitoring -l app=prometheus

# Check Grafana
kubectl get pods -n monitoring -l app.kubernetes.io/name=grafana

# Check Loki
kubectl get pods -n monitoring -l app=loki

# Check ServiceMonitors
kubectl get servicemonitors -n velure

# Check PrometheusRules
kubectl get prometheusrules -n monitoring
```

#### Access Grafana

```bash
# Port forward
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Get admin password
kubectl get secret -n monitoring kube-prometheus-stack-grafana \
  -o jsonpath="{.data.admin-password}" | base64 --decode
```

Or access via LoadBalancer (if configured):

```bash
kubectl get svc -n monitoring kube-prometheus-stack-grafana
```

---

## Dashboards

### Pre-configured Dashboards

All dashboards are automatically provisioned in Grafana:

#### Infrastructure

1. **Kubernetes Cluster Overview** (`velure-k8s-overview`)
   - Total nodes, pods, deployments, services
   - CPU and memory usage by pod
   - Pod status table

2. **Node Metrics** (`velure-node-metrics`)
   - CPU, memory, disk, network I/O per node
   - Resource utilization trends

#### Application Services

3. **Velure Services Overview** (`velure-services`)
   - Request rate across all services
   - Latency (p95, p99)
   - Error rate percentage
   - Business metrics (total users, products)
   - Cache hit rates
   - Login attempts

#### Databases

4. **PostgreSQL Metrics** (`velure-postgres`)
   - Active connections
   - Cache hit ratio
   - Transactions per second
   - Tuple operations (inserts, updates, deletes)

5. **MongoDB Metrics** (`velure-mongodb`)
   - Current connections
   - Operations per second (insert, query, update, delete)
   - Memory usage
   - Database size

6. **Redis Metrics** (`velure-redis`)
   - Connected clients
   - Hit rate percentage
   - Total keys
   - Memory usage
   - Commands per second
   - Cache hits vs misses

#### Messaging

7. **RabbitMQ Metrics** (`velure-rabbitmq`)
   - Total connections and channels
   - Messages in queues
   - Memory usage
   - Publish vs delivery rate
   - Queue depth

#### Logs

8. **Logs Dashboard** (`velure-logs`)
   - Real-time log streaming
   - Logs by level (ERROR, WARN, INFO, DEBUG)
   - Logs by service
   - Error logs only view
   - Search and filtering

### Importing Dashboards

Dashboards are automatically loaded from:
- **Local**: `/infrastructure/local/monitoring/grafana/dashboards/*.json`
- **Kubernetes**: Via ConfigMaps and Grafana provisioning

---

## Alerts

### Alert Categories

#### Application Alerts

| Alert | Severity | Threshold | Description |
|-------|----------|-----------|-------------|
| `HighErrorRate` | Warning | >1% | Service error rate above 1% for 5 min |
| `CriticalErrorRate` | Critical | >5% | Service error rate above 5% for 2 min |
| `HighLatency` | Warning | >1s | p99 latency above 1 second |
| `ServiceDown` | Critical | N/A | Service unreachable for 1 min |

#### Infrastructure Alerts

| Alert | Severity | Threshold | Description |
|-------|----------|-----------|-------------|
| `HighCPUUsage` | Warning | >80% | Pod CPU usage above 80% for 10 min |
| `HighMemoryUsage` | Warning | >80% | Pod memory usage above 80% for 10 min |
| `PodCrashLooping` | Critical | N/A | Pod restarting repeatedly |
| `DiskSpaceLow` | Warning | <20% | Disk space below 20% |

#### Database Alerts

| Alert | Severity | Threshold | Description |
|-------|----------|-----------|-------------|
| `PostgreSQLConnectionPoolHigh` | Warning | >80 | Too many active connections |
| `PostgreSQLCacheHitRatioLow` | Warning | <90% | Cache efficiency degraded |
| `MongoDBReplicationLag` | Warning | >10s | Replication lag detected |
| `RedisMemoryHigh` | Warning | >1GB | Redis memory usage high |

#### Messaging Alerts

| Alert | Severity | Threshold | Description |
|-------|----------|-----------|-------------|
| `RabbitMQQueueGrowing` | Warning | N/A | Queue growing for 10 min |
| `RabbitMQHighMessageCount` | Warning | >10k | Too many messages queued |
| `RabbitMQNoConsumers` | Critical | 0 | No consumers on queue |

### Configuring Alert Receivers

Edit AlertManager configuration to add Slack, Email, or PagerDuty:

```bash
# Kubernetes
kubectl edit alertmanagerconfig -n monitoring kube-prometheus-stack-alertmanager

# Add receivers
receivers:
  - name: 'slack-notifications'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
        title: 'Velure Alert'
        text: '{{ .CommonAnnotations.description }}'
```

---

## Accessing Services

### Local (Docker Compose)

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| AlertManager | http://localhost:9093 | - |
| Loki | http://localhost:3100 | - |

### Kubernetes (Port Forward)

```bash
# Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Prometheus
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090

# AlertManager
kubectl port-forward -n monitoring svc/kube-prometheus-stack-alertmanager 9093:9093

# Loki
kubectl port-forward -n monitoring svc/loki 3100:3100
```

---

## Troubleshooting

### Common Issues

#### Prometheus not scraping targets

```bash
# Check ServiceMonitors
kubectl get servicemonitors -n velure

# Check Prometheus targets
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# Visit: http://localhost:9090/targets

# Check service endpoints
kubectl get endpoints -n velure
```

#### Grafana dashboards not loading

```bash
# Check datasources
kubectl exec -n monitoring deployment/kube-prometheus-stack-grafana -- \
  grafana-cli admin data-sources ls

# Re-provision datasources
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n monitoring
```

#### Loki not receiving logs

```bash
# Check Promtail pods
kubectl get pods -n monitoring -l app=promtail

# Check Promtail logs
kubectl logs -n monitoring -l app=promtail --tail=100

# Verify Loki is reachable
kubectl exec -n monitoring -it <promtail-pod> -- \
  wget -O- http://loki:3100/ready
```

#### High memory usage

```bash
# Reduce Prometheus retention
# Edit prometheus-values.yaml:
retention: 7d  # instead of 15d

# Reduce Loki retention
# Edit loki-stack-values.yaml:
retention_period: 7d  # instead of 30d

# Apply changes
helm upgrade kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  -f kube-prometheus-stack-values.yaml -n monitoring
```

### Debugging

```bash
# Check logs
kubectl logs -n monitoring -l app=prometheus --tail=100
kubectl logs -n monitoring -l app.kubernetes.io/name=grafana --tail=100
kubectl logs -n monitoring -l app=loki --tail=100

# Describe resources
kubectl describe pod -n monitoring <pod-name>
kubectl describe servicemonitor -n velure <servicemonitor-name>

# Check CRDs
kubectl get prometheusrules -n monitoring
kubectl get servicemonitors --all-namespaces
```

---

## Useful Commands

### Metrics

```bash
# Query Prometheus
curl -s 'http://localhost:9090/api/v1/query?query=up' | jq

# Get recording rules
kubectl get prometheusrules -n monitoring velure-recording-rules -o yaml

# Check targets
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

### Logs

```bash
# Query Loki
curl -G -s 'http://localhost:3100/loki/api/v1/query' \
  --data-urlencode 'query={service_type="auth"}' | jq

# Stream logs
curl -G -s 'http://localhost:3100/loki/api/v1/tail' \
  --data-urlencode 'query={service_type="auth"}' | jq

# Check Loki metrics
curl -s http://localhost:3100/metrics | grep loki_
```

### Alerts

```bash
# List active alerts
curl -s http://localhost:9090/api/v1/alerts | jq

# Check alert rules
kubectl get prometheusrules -n monitoring velure-alerts -o yaml

# Silence alert
kubectl exec -n monitoring alertmanager-kube-prometheus-stack-alertmanager-0 -- \
  amtool silence add alertname="HighErrorRate" --duration=1h
```

---

## Maintenance

### Backup

```bash
# Backup Grafana dashboards
kubectl get configmaps -n monitoring -l grafana_dashboard=1 -o yaml > grafana-dashboards-backup.yaml

# Backup Prometheus data
kubectl exec -n monitoring prometheus-kube-prometheus-stack-prometheus-0 -- \
  tar -czf /tmp/prometheus-data.tar.gz /prometheus

# Copy backup
kubectl cp monitoring/prometheus-kube-prometheus-stack-prometheus-0:/tmp/prometheus-data.tar.gz ./prometheus-backup.tar.gz
```

### Upgrades

```bash
# Update Helm repos
helm repo update

# Upgrade prometheus-stack
helm upgrade kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  -f kube-prometheus-stack-values.yaml -n monitoring

# Upgrade loki
helm upgrade loki grafana/loki-stack \
  -f loki-stack-values.yaml -n monitoring
```

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [PromQL Cheat Sheet](https://promlabs.com/promql-cheat-sheet/)

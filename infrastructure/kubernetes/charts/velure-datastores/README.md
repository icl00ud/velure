# Velure Datastores Helm Chart

Unified chart that deploys MongoDB, Redis, and RabbitMQ for the Velure environment.

## ⚠️ Important: chart has dependencies

**This is an umbrella chart** that depends on Bitnami charts. Before installing you **MUST** download the dependencies:

```bash
# Option 1: build (downloads and stores under charts/)
helm dependency build infrastructure/kubernetes/charts/velure-datastores

# Option 2: update (downloads and refreshes every time)
helm dependency update infrastructure/kubernetes/charts/velure-datastores

# Then install normally
helm upgrade --install velure-datastores infrastructure/kubernetes/charts/velure-datastores -n datastores
```

**External dependencies (Bitnami):**
- `mongodb`: v14.12.1
- `redis`: v18.19.1
- `rabbitmq`: v13.0.1

**⚠️ Skipping `helm dependency build/update` will produce:**
```
Error: found in Chart.yaml, but missing in charts/ directory: mongodb, redis, rabbitmq
```

---

## ⚠️ Environments

**Suitable for development and test environments.** For production, prefer managed services:
- **MongoDB** → Amazon DocumentDB
- **Redis** → Amazon ElastiCache
- **RabbitMQ** → Amazon MQ

## Components

### MongoDB
- **Use**: Product Service (product catalog)
- **Database**: `productdb`
- **User**: `productuser`
- **Port**: 27017

### Redis
- **Use**: caching (Product Service, Auth Service)
- **Port**: 6379

### RabbitMQ
- **Use**: message broker between publish-order and process-order services
- **Ports**: 5672 (AMQP), 15672 (Management UI)
- **Exchanges**: `orders` (topic)
- **Queues**:
  - `process-order-queue`
  - `publish-order-status-updates`

## Install

```bash
# 1. Add the Bitnami repository
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# 2. Download chart dependencies (REQUIRED!)
helm dependency build infrastructure/kubernetes/charts/velure-datastores

# 3. Create the namespace
kubectl create namespace datastores

# 4. Install the chart
helm install velure-datastores infrastructure/kubernetes/charts/velure-datastores \
  --namespace datastores \
  --create-namespace

# Or with custom values
helm install velure-datastores infrastructure/kubernetes/charts/velure-datastores \
  --namespace datastores \
  --values custom-values.yaml
```

**⚠️ IMPORTANT:** never skip step 2 (`helm dependency build`). The install will fail without it.

## Verify

```bash
# Pods
kubectl get pods -n datastores

# Services
kubectl get svc -n datastores

# PVCs
kubectl get pvc -n datastores

# MongoDB logs
kubectl logs -n datastores -l app.kubernetes.io/name=mongodb

# Redis logs
kubectl logs -n datastores -l app.kubernetes.io/name=redis

# RabbitMQ logs
kubectl logs -n datastores -l app.kubernetes.io/name=rabbitmq
```

## Service Access

### MongoDB
```bash
# Internal connection string
mongodb://productuser:product_password@velure-datastores-mongodb:27017/productdb

# Port-forward for local access
kubectl port-forward -n datastores svc/velure-datastores-mongodb 27017:27017

# Connect via CLI
kubectl exec -it -n datastores velure-datastores-mongodb-0 -- mongosh -u productuser -p product_password productdb
```

### Redis
```bash
# Internal connection string
redis://:redis_password@velure-datastores-redis-master:6379

# Port-forward
kubectl port-forward -n datastores svc/velure-datastores-redis-master 6379:6379

# Connect via CLI
kubectl exec -it -n datastores velure-datastores-redis-master-0 -- redis-cli -a redis_password
```

### RabbitMQ
```bash
# Internal AMQP URL
amqp://publisher-order:publisher_password@velure-datastores-rabbitmq:5672/

# Management UI port-forward
kubectl port-forward -n datastores svc/velure-datastores-rabbitmq 15672:15672

# Open: http://localhost:15672
# User: admin
# Password: admin_password
```

## Configuration

### Disable components

```yaml
# values.yaml
mongodb:
  enabled: false  # disable MongoDB

redis:
  enabled: false  # disable Redis

rabbitmq:
  enabled: false  # disable RabbitMQ
```

### Adjust resources

```yaml
mongodb:
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 512Mi
```

### Persistence

```yaml
mongodb:
  persistence:
    enabled: true
    storageClass: "gp3"  # or "standard" for development
    size: 20Gi
```

## Troubleshooting

### MongoDB does not start
```bash
# Check logs
kubectl logs -n datastores -l app.kubernetes.io/name=mongodb

# Check PVC
kubectl get pvc -n datastores

# Describe pod
kubectl describe pod -n datastores -l app.kubernetes.io/name=mongodb
```

### RabbitMQ — queues are not created
```bash
# Check whether definitions loaded
kubectl exec -n datastores velure-datastores-rabbitmq-0 -- rabbitmqctl list_queues

# Recreate the load-definition secret
kubectl delete secret -n datastores rabbitmq-load-definition
helm upgrade velure-datastores ./velure-datastores -n datastores
```

### Redis — connection refused
```bash
# Check pod state
kubectl get pods -n datastores -l app.kubernetes.io/name=redis

# Test the connection
kubectl exec -it -n datastores velure-datastores-redis-master-0 -- redis-cli ping
```

## Backup and Restore

### MongoDB backup
```bash
# Create a backup
kubectl exec -n datastores velure-datastores-mongodb-0 -- \
  mongodump --uri="mongodb://productuser:product_password@localhost:27017/productdb" \
  --out=/tmp/backup

# Copy it locally
kubectl cp datastores/velure-datastores-mongodb-0:/tmp/backup ./mongodb-backup
```

### Redis backup
```bash
# Trigger BGSAVE
kubectl exec -n datastores velure-datastores-redis-master-0 -- redis-cli -a redis_password BGSAVE

# Copy the RDB file
kubectl cp datastores/velure-datastores-redis-master-0:/data/dump.rdb ./redis-backup.rdb
```

## Uninstall

```bash
# Remove the chart (PVCs are kept)
helm uninstall velure-datastores -n datastores

# Remove PVCs too
kubectl delete pvc -n datastores --all

# Remove the namespace
kubectl delete namespace datastores
```

## Monitoring

All components export Prometheus metrics when `metrics.enabled: true`:
- MongoDB: `:9216/metrics`
- Redis: `:9121/metrics`
- RabbitMQ: `:15692/metrics`

ServiceMonitors are created automatically for scraping by the Prometheus Operator.

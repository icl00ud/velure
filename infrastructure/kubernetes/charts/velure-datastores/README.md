# Velure Datastores Helm Chart

Este chart unificado implanta MongoDB, Redis e RabbitMQ para o ambiente Velure.

## ⚠️ Importante: Chart com Dependências

**Este é um umbrella chart** que depende de charts do Bitnami. Antes de instalar, você **DEVE** baixar as dependências:

```bash
# Opção 1: Build (baixa e salva em charts/)
helm dependency build infrastructure/kubernetes/charts/velure-datastores

# Opção 2: Update (baixa e atualiza sempre)
helm dependency update infrastructure/kubernetes/charts/velure-datastores

# Depois instalar normalmente
helm upgrade --install velure-datastores infrastructure/kubernetes/charts/velure-datastores -n datastores
```

**Dependências externas (Bitnami):**
- `mongodb`: v14.12.1
- `redis`: v18.19.1
- `rabbitmq`: v13.0.1

**⚠️ Sem executar `helm dependency build/update` primeiro, você receberá o erro:**
```
Error: found in Chart.yaml, but missing in charts/ directory: mongodb, redis, rabbitmq
```

---

## ⚠️ Ambientes

**Este chart é adequado para ambientes de desenvolvimento e teste.** Para produção, recomenda-se usar serviços gerenciados:
- **MongoDB** → Amazon DocumentDB
- **Redis** → Amazon ElastiCache
- **RabbitMQ** → Amazon MQ

## Componentes

### MongoDB
- **Uso**: Product Service (catálogo de produtos)
- **Database**: `productdb`
- **User**: `productuser`
- **Porta**: 27017

### Redis
- **Uso**: Caching (Product Service, Auth Service)
- **Porta**: 6379

### RabbitMQ
- **Uso**: Message broker entre publish-order e process-order services
- **Porta**: 5672 (AMQP), 15672 (Management UI)
- **Exchanges**: `orders` (topic)
- **Queues**:
  - `process-order-queue`
  - `publish-order-status-updates`

## Instalação

```bash
# 1. Adicionar repositório Bitnami
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# 2. Baixar dependências do chart (OBRIGATÓRIO!)
helm dependency build infrastructure/kubernetes/charts/velure-datastores

# 3. Criar namespace
kubectl create namespace datastores

# 4. Instalar chart
helm install velure-datastores infrastructure/kubernetes/charts/velure-datastores \
  --namespace datastores \
  --create-namespace

# Ou usar values customizados
helm install velure-datastores infrastructure/kubernetes/charts/velure-datastores \
  --namespace datastores \
  --values custom-values.yaml
```

**⚠️ IMPORTANTE:** Nunca pule o passo 2 (`helm dependency build`). Sem ele, a instalação falhará.

## Verificação

```bash
# Ver pods
kubectl get pods -n datastores

# Ver services
kubectl get svc -n datastores

# Ver PVCs
kubectl get pvc -n datastores

# Logs do MongoDB
kubectl logs -n datastores -l app.kubernetes.io/name=mongodb

# Logs do Redis
kubectl logs -n datastores -l app.kubernetes.io/name=redis

# Logs do RabbitMQ
kubectl logs -n datastores -l app.kubernetes.io/name=rabbitmq
```

## Acesso aos Serviços

### MongoDB
```bash
# Connection string interno
mongodb://productuser:product_password@velure-datastores-mongodb:27017/productdb

# Port-forward para acesso local
kubectl port-forward -n datastores svc/velure-datastores-mongodb 27017:27017

# Conectar via CLI
kubectl exec -it -n datastores velure-datastores-mongodb-0 -- mongosh -u productuser -p product_password productdb
```

### Redis
```bash
# Connection string interno
redis://:redis_password@velure-datastores-redis-master:6379

# Port-forward
kubectl port-forward -n datastores svc/velure-datastores-redis-master 6379:6379

# Conectar via CLI
kubectl exec -it -n datastores velure-datastores-redis-master-0 -- redis-cli -a redis_password
```

### RabbitMQ
```bash
# AMQP URL interno
amqp://publisher-order:publisher_password@velure-datastores-rabbitmq:5672/

# Management UI port-forward
kubectl port-forward -n datastores svc/velure-datastores-rabbitmq 15672:15672

# Acessar: http://localhost:15672
# User: admin
# Password: admin_password
```

## Configuração

### Desabilitar Componentes

```yaml
# values.yaml
mongodb:
  enabled: false  # Desabilita MongoDB

redis:
  enabled: false  # Desabilita Redis

rabbitmq:
  enabled: false  # Desabilita RabbitMQ
```

### Alterar Recursos

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

### Persistência

```yaml
mongodb:
  persistence:
    enabled: true
    storageClass: "gp3"  # Ou "standard" para development
    size: 20Gi
```

## Troubleshooting

### MongoDB não inicia
```bash
# Verificar logs
kubectl logs -n datastores -l app.kubernetes.io/name=mongodb

# Verificar PVC
kubectl get pvc -n datastores

# Descrever pod
kubectl describe pod -n datastores -l app.kubernetes.io/name=mongodb
```

### RabbitMQ - Queues não criadas
```bash
# Verificar se definitions foram carregadas
kubectl exec -n datastores velure-datastores-rabbitmq-0 -- rabbitmqctl list_queues

# Recriar load definition secret
kubectl delete secret -n datastores rabbitmq-load-definition
helm upgrade velure-datastores ./velure-datastores -n datastores
```

### Redis - Connection refused
```bash
# Verificar se está rodando
kubectl get pods -n datastores -l app.kubernetes.io/name=redis

# Testar conexão
kubectl exec -it -n datastores velure-datastores-redis-master-0 -- redis-cli ping
```

## Backup e Restore

### MongoDB Backup
```bash
# Criar backup
kubectl exec -n datastores velure-datastores-mongodb-0 -- \
  mongodump --uri="mongodb://productuser:product_password@localhost:27017/productdb" \
  --out=/tmp/backup

# Copiar para local
kubectl cp datastores/velure-datastores-mongodb-0:/tmp/backup ./mongodb-backup
```

### Redis Backup
```bash
# Trigger BGSAVE
kubectl exec -n datastores velure-datastores-redis-master-0 -- redis-cli -a redis_password BGSAVE

# Copiar RDB file
kubectl cp datastores/velure-datastores-redis-master-0:/data/dump.rdb ./redis-backup.rdb
```

## Desinstalação

```bash
# Remover chart (mantém PVCs)
helm uninstall velure-datastores -n datastores

# Remover PVCs também
kubectl delete pvc -n datastores --all

# Remover namespace
kubectl delete namespace datastores
```

## Monitoramento

Todos os componentes exportam métricas Prometheus quando `metrics.enabled: true`:
- MongoDB: `:9216/metrics`
- Redis: `:9121/metrics`
- RabbitMQ: `:15692/metrics`

ServiceMonitors são criados automaticamente para scraping pelo Prometheus Operator.

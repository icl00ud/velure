# Deploy Kubernetes (Helm)

Este diretório contém charts Helm para tornar o projeto Kubernetes‑native. Ele contempla:

- velure-auth (já existente)
- velure-product (já existente)
- velure-ui (já existente)
- velure-publish-order (novo)
- velure-process-order (novo)
- velure-rabbitmq (novo)

Além disso, arquivos de Secret de exemplo para MongoDB e Redis estão em `manifests/`.

## Pré‑requisitos

- Kubernetes 1.26+
- Helm 3.12+
- Namespaces: `authentication`, `order`, `frontend`, `database`
- PostgreSQL/MongoDB/Redis no cluster (pode usar charts oficiais) ou configure conforme seus providers gerenciados
- Opcional: Prometheus/ServiceMonitor (kube‑prometheus‑stack) para métricas

## Namespaces

```sh
kubectl create ns authentication || true
kubectl create ns order || true
kubectl create ns frontend || true
kubectl create ns database || true
```

## Segredos de exemplo

- Banco de pedidos (Postgres) e RabbitMQ (Order domain):

```sh
kubectl -n order create secret generic order-database \
  --from-literal=url="postgres://user:password@postgres.database.svc.cluster.local:5432/orders?sslmode=disable" --dry-run=client -o yaml | kubectl apply -f -

kubectl -n order create secret generic rabbitmq-conn \
  --from-literal=url="amqp://publisher:publisher@rabbitmq.order.svc.cluster.local:5672/" --dry-run=client -o yaml | kubectl apply -f -
```

- Auth Service (JWT/Session/Postgres):

```sh
kubectl -n authentication create secret generic velure-auth-jwt \
  --from-literal=secret="change-me" \
  --from-literal=expiresIn="1h" \
  --from-literal=refreshSecret="change-me-refresh" \
  --from-literal=refreshExpiresIn="7d" --dry-run=client -o yaml | kubectl apply -f -

kubectl -n authentication create secret generic velure-auth-session \
  --from-literal=secret="session-secret" \
  --from-literal=expiresIn="86400000" --dry-run=client -o yaml | kubectl apply -f -

kubectl -n authentication create secret generic velure-auth-postgres \
  --from-literal=url="postgres://user:password@postgres.database.svc.cluster.local:5432/auth?sslmode=disable" --dry-run=client -o yaml | kubectl apply -f -
```

- Mongo/Redis (Product): use `kubernetes/manifests/*` ou crie secrets equivalentes.

## Build das imagens

As imagens Docker são multi‑stage e já preparadas para rodar como não‑root. Gere e publique em seu registry:

```sh
# Exemplo (ajuste tags/repo)
DOCKER_BUILDKIT=1 docker build -t <repo>/velure-auth-service:latest ./auth-service
DOCKER_BUILDKIT=1 docker build -t <repo>/velure-product-service:latest ./product-service
DOCKER_BUILDKIT=1 docker build -t <repo>/velure-publish-order-service:latest ./publish-order-service
DOCKER_BUILDKIT=1 docker build -t <repo>/velure-process-order-service:latest ./process-order-service
DOCKER_BUILDKIT=1 docker build -t <repo>/velure-ui-service:latest ./ui-service
```

Atualize os `values.yaml` dos charts com seus repositórios/tags.

## Instalação (Helm)

```sh
# Auth
helm upgrade --install velure-auth ./kubernetes/charts/velure-auth -n authentication \
  --set image.repository=<repo>/velure-auth-service --set image.tag=latest

# Product
helm upgrade --install velure-product ./kubernetes/charts/velure-product -n order \
  --set image.repository=<repo>/velure-product-service --set image.tag=latest

# UI
helm upgrade --install velure-ui ./kubernetes/charts/velure-ui -n frontend \
  --set image.repository=<repo>/velure-ui-service --set image.tag=latest

# RabbitMQ (opcional, caso não use gerenciado)
helm upgrade --install velure-rabbitmq ./kubernetes/charts/velure-rabbitmq -n order

# Publish Order
helm upgrade --install velure-publish-order ./kubernetes/charts/velure-publish-order -n order \
  --set image.repository=<repo>/velure-publish-order-service --set image.tag=latest

# Process Order
helm upgrade --install velure-process-order ./kubernetes/charts/velure-process-order -n order \
  --set image.repository=<repo>/velure-process-order-service --set image.tag=latest
```

## Saúde e Probes

- auth-service: `GET /health`
- product-service: `GET /health`
- publish-order-service: `GET /health`
- process-order-service: `GET /health`

## Observabilidade

- Auth expõe métricas Prometheus em `/authentication/authMetrics`.
- Para integrar com Prometheus Operator, adicione ServiceMonitor conforme sua stack.

## Notas

- Os charts usam Secrets para strings de conexão. Ajuste os nomes dos Secrets conforme seu ambiente.
- Para usar External Secrets (AWS Secrets Manager, etc.), adapte os templates ou crie `ExternalSecret` no namespace adequado.
